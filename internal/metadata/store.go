package metadata

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"pdfmeta/internal/filesafe"
	"pdfmeta/internal/model"
	"pdfmeta/internal/pdf"
	"pdfmeta/internal/xmp"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

var _ model.MetadataStore = (*Store)(nil)

var (
	xrefStreamPattern   = regexp.MustCompile(`(?s)\b\d+\s+\d+\s+obj\b.*?/Type\s*/XRef\b`)
	xrefStmKeyPattern   = regexp.MustCompile(`/XRefStm\s+\d+\b`)
	objectStreamPattern = regexp.MustCompile(`(?s)\b\d+\s+\d+\s+obj\b.*?/Type\s*/ObjStm\b`)
)

func (s *Store) Read(ctx context.Context, inputPath string) (model.MetadataReadResult, error) {
	if err := ctxErr(ctx); err != nil {
		return model.MetadataReadResult{}, err
	}

	doc, err := pdf.Open(inputPath)
	if err != nil {
		return model.MetadataReadResult{}, err
	}

	meta, infoFound, xmpFound := readNativeMetadata(doc.Bytes())
	return model.MetadataReadResult{
		Encrypted:  doc.Encrypted(),
		Metadata:   meta,
		InfoFound:  infoFound,
		XMPFound:   xmpFound,
		Normalized: false,
	}, nil
}

func (s *Store) Write(ctx context.Context, req model.MetadataWriteRequest) (model.MetadataReadResult, error) {
	if err := ctxErr(ctx); err != nil {
		return model.MetadataReadResult{}, err
	}
	if req.InputPath == "" {
		return model.MetadataReadResult{}, &model.AppError{Code: model.ErrValidation, Message: "input path is required"}
	}
	dst, err := writeTarget(req)
	if err != nil {
		return model.MetadataReadResult{}, err
	}

	doc, err := pdf.Open(req.InputPath)
	if err != nil {
		return model.MetadataReadResult{}, err
	}
	if doc.Encrypted() {
		return model.MetadataReadResult{}, &model.AppError{Code: model.ErrPDFEncrypted, Message: "cannot write encrypted pdf"}
	}
	if err := unsupportedStructureError(doc.Bytes()); err != nil {
		return model.MetadataReadResult{}, err
	}

	current, _, _ := readNativeMetadata(doc.Bytes())
	next := applyPatch(current, req.Set)
	next = applyUnset(next, req.Unset, req.UnsetAll)

	xmpPacket, err := xmp.Marshal(next)
	if err != nil {
		return model.MetadataReadResult{}, &model.AppError{Code: model.ErrInternal, Message: "encode xmp packet", Cause: err}
	}

	updated, err := writeNativeIncremental(doc.Bytes(), next, xmpPacket)
	if err != nil {
		return model.MetadataReadResult{}, err
	}

	if err := filesafe.WriteAtomic(dst, updated, 0o644); err != nil {
		return model.MetadataReadResult{}, &model.AppError{Code: model.ErrIO, Message: fmt.Sprintf("write %q", dst), Cause: err}
	}

	return model.MetadataReadResult{
		Encrypted:  false,
		Metadata:   next,
		InfoFound:  true,
		XMPFound:   true,
		Normalized: false,
	}, nil
}

func ctxErr(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return &model.AppError{Code: model.ErrInternal, Message: "operation canceled", Cause: ctx.Err()}
	default:
		return nil
	}
}

func writeTarget(req model.MetadataWriteRequest) (string, error) {
	if req.InPlace {
		return req.InputPath, nil
	}
	if req.OutputPath == "" {
		return "", &model.AppError{Code: model.ErrValidation, Message: "output path is required when not writing in place"}
	}
	return req.OutputPath, nil
}

func applyPatch(cur model.Metadata, patch model.MetadataPatch) model.Metadata {
	next := cur
	if patch.Title != nil {
		next.Title = *patch.Title
	}
	if patch.Author != nil {
		next.Author = *patch.Author
	}
	if patch.Subject != nil {
		next.Subject = *patch.Subject
	}
	if patch.Keywords != nil {
		next.Keywords = *patch.Keywords
	}
	if patch.Creator != nil {
		next.Creator = *patch.Creator
	}
	if patch.Producer != nil {
		next.Producer = *patch.Producer
	}
	if patch.CreationDate != nil {
		next.CreationDate = *patch.CreationDate
	}
	if patch.ModDate != nil {
		next.ModDate = *patch.ModDate
	}
	return next
}

func applyUnset(cur model.Metadata, fields []model.Field, unsetAll bool) model.Metadata {
	if unsetAll {
		return model.Metadata{}
	}
	next := cur
	for _, f := range fields {
		switch f {
		case model.FieldTitle:
			next.Title = ""
		case model.FieldAuthor:
			next.Author = ""
		case model.FieldSubject:
			next.Subject = ""
		case model.FieldKeywords:
			next.Keywords = ""
		case model.FieldCreator:
			next.Creator = ""
		case model.FieldProducer:
			next.Producer = ""
		case model.FieldCreationDate:
			next.CreationDate = ""
		case model.FieldModDate:
			next.ModDate = ""
		}
	}
	return next
}

type objRef struct {
	Obj int
	Gen int
}

func readNativeMetadata(b []byte) (model.Metadata, bool, bool) {
	rootRef, infoRef, ok := parseTrailerRefs(b)
	if !ok {
		return model.Metadata{}, false, false
	}

	meta := model.Metadata{}
	infoFound := false
	xmpFound := false

	if infoRef.Obj > 0 {
		if body, ok := objectBody(b, infoRef.Obj, infoRef.Gen); ok {
			if dict, ok := firstDict(body); ok {
				meta = mergeMetadata(parseInfoDict(dict), meta)
				infoFound = true
			}
		}
	}

	if rootRef.Obj > 0 {
		if rootBody, ok := objectBody(b, rootRef.Obj, rootRef.Gen); ok {
			if rootDict, ok := firstDict(rootBody); ok {
				if mdRef, ok := parseNamedRef(rootDict, "Metadata"); ok {
					if mdBody, ok := objectBody(b, mdRef.Obj, mdRef.Gen); ok {
						if stream, ok := streamContent(mdBody); ok {
							if x, err := xmp.Unmarshal(stream); err == nil {
								meta = mergeMetadata(meta, x)
								xmpFound = true
							}
						}
					}
				}
			}
		}
	}

	return meta, infoFound, xmpFound
}

func unsupportedStructureError(b []byte) error {
	switch {
	case xrefStreamPattern.Match(b) || xrefStmKeyPattern.Match(b):
		return &model.AppError{
			Code:    model.ErrPDFMalformed,
			Message: "cannot rewrite PDFs that use xref streams yet",
		}
	case objectStreamPattern.Match(b):
		return &model.AppError{
			Code:    model.ErrPDFMalformed,
			Message: "cannot rewrite PDFs that use object streams yet",
		}
	default:
		return nil
	}
}

func writeNativeIncremental(src []byte, meta model.Metadata, xmpPacket []byte) ([]byte, error) {
	rootRef, _, ok := parseTrailerRefs(src)
	if !ok {
		return nil, &model.AppError{Code: model.ErrPDFMalformed, Message: "could not parse trailer root reference"}
	}
	startXRef, ok := parseStartXRef(src)
	if !ok {
		return nil, &model.AppError{Code: model.ErrPDFMalformed, Message: "could not parse startxref"}
	}

	maxObj := maxObjectNumber(src)
	if maxObj < 1 {
		return nil, &model.AppError{Code: model.ErrPDFMalformed, Message: "could not detect object numbers"}
	}

	rootBody, ok := objectBody(src, rootRef.Obj, rootRef.Gen)
	if !ok {
		return nil, &model.AppError{Code: model.ErrPDFMalformed, Message: "could not read catalog object"}
	}
	rootDict, ok := firstDict(rootBody)
	if !ok {
		return nil, &model.AppError{Code: model.ErrPDFMalformed, Message: "catalog dictionary missing"}
	}

	infoObj := maxObj + 1
	metadataObj := maxObj + 2
	catalogObj := maxObj + 3

	newCatalogDict := upsertNamedRef(rootDict, "Metadata", objRef{Obj: metadataObj, Gen: 0})

	infoObject := renderInfoObject(infoObj, meta)
	metadataObject := renderMetadataObject(metadataObj, xmpPacket)
	catalogObject := renderCatalogObject(catalogObj, newCatalogDict)

	out := append([]byte(nil), src...)
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, '\n')
	}

	offInfo := len(out)
	out = append(out, infoObject...)
	offMetadata := len(out)
	out = append(out, metadataObject...)
	offCatalog := len(out)
	out = append(out, catalogObject...)

	xrefOffset := len(out)
	xref := renderXRef(infoObj, []int{offInfo, offMetadata, offCatalog})
	out = append(out, xref...)

	trailer := renderTrailer(catalogObj+1, objRef{Obj: catalogObj, Gen: 0}, objRef{Obj: infoObj, Gen: 0}, startXRef)
	out = append(out, trailer...)
	out = append(out, []byte("startxref\n")...)
	out = append(out, []byte(strconv.Itoa(xrefOffset)+"\n")...)
	out = append(out, []byte("%%EOF\n")...)

	return out, nil
}

func parseTrailerRefs(b []byte) (objRef, objRef, bool) {
	trailer := lastTrailerDict(b)
	if trailer == "" {
		return objRef{}, objRef{}, false
	}
	root, ok := parseNamedRef(trailer, "Root")
	if !ok {
		return objRef{}, objRef{}, false
	}
	info, _ := parseNamedRef(trailer, "Info")
	return root, info, true
}

func lastTrailerDict(b []byte) string {
	s := string(b)
	idx := strings.LastIndex(s, "trailer")
	if idx < 0 {
		return ""
	}
	tail := s[idx+len("trailer"):]
	start := strings.Index(tail, "<<")
	if start < 0 {
		return ""
	}
	absStart := idx + len("trailer") + start
	end, ok := matchDictEnd(s, absStart)
	if !ok {
		return ""
	}
	return s[absStart : end+2]
}

func matchDictEnd(s string, start int) (int, bool) {
	depth := 0
	for i := start; i < len(s)-1; i++ {
		if s[i] == '<' && s[i+1] == '<' {
			depth++
			i++
			continue
		}
		if s[i] == '>' && s[i+1] == '>' {
			depth--
			if depth == 0 {
				return i, true
			}
			i++
		}
	}
	return 0, false
}

func parseNamedRef(dict, name string) (objRef, bool) {
	re := regexp.MustCompile(`/` + regexp.QuoteMeta(name) + `\s+(\d+)\s+(\d+)\s+R`)
	m := re.FindStringSubmatch(dict)
	if len(m) != 3 {
		return objRef{}, false
	}
	o, _ := strconv.Atoi(m[1])
	g, _ := strconv.Atoi(m[2])
	return objRef{Obj: o, Gen: g}, true
}

func parseStartXRef(b []byte) (int, bool) {
	s := string(b)
	idx := strings.LastIndex(s, "startxref")
	if idx < 0 {
		return 0, false
	}
	rest := strings.TrimSpace(s[idx+len("startxref"):])
	if rest == "" {
		return 0, false
	}
	line := rest
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		line = strings.TrimSpace(rest[:nl])
	}
	n, err := strconv.Atoi(line)
	if err != nil {
		return 0, false
	}
	return n, true
}

func maxObjectNumber(b []byte) int {
	re := regexp.MustCompile(`(?m)(\d+)\s+(\d+)\s+obj\b`)
	all := re.FindAllSubmatch(b, -1)
	max := 0
	for _, m := range all {
		n, err := strconv.Atoi(string(m[1]))
		if err == nil && n > max {
			max = n
		}
	}
	return max
}

func objectBody(b []byte, obj, gen int) (string, bool) {
	pattern := fmt.Sprintf(`(?s)(^|[\r\n])%d\s+%d\s+obj\b(.*?)\bendobj`, obj, gen)
	re := regexp.MustCompile(pattern)
	m := re.FindSubmatch(b)
	if len(m) != 3 {
		return "", false
	}
	return string(m[2]), true
}

func firstDict(body string) (string, bool) {
	start := strings.Index(body, "<<")
	if start < 0 {
		return "", false
	}
	end, ok := matchDictEnd(body, start)
	if !ok {
		return "", false
	}
	return body[start : end+2], true
}

func streamContent(objBody string) ([]byte, bool) {
	i := strings.Index(objBody, "stream")
	if i < 0 {
		return nil, false
	}
	j := strings.Index(objBody[i+6:], "endstream")
	if j < 0 {
		return nil, false
	}
	start := i + 6
	for start < len(objBody) && (objBody[start] == '\n' || objBody[start] == '\r') {
		start++
	}
	end := i + 6 + j
	content := strings.TrimRight(objBody[start:end], "\r\n")
	return []byte(content), true
}

func parseInfoDict(dict string) model.Metadata {
	get := func(key string) string {
		return decodePDFString(findDictValue(dict, key))
	}
	return model.Metadata{
		Title:        get("Title"),
		Author:       get("Author"),
		Subject:      get("Subject"),
		Keywords:     get("Keywords"),
		Creator:      get("Creator"),
		Producer:     get("Producer"),
		CreationDate: get("CreationDate"),
		ModDate:      get("ModDate"),
	}
}

func findDictValue(dict, key string) string {
	for i := 0; i < len(dict); {
		switch dict[i] {
		case '(':
			end, ok := scanPDFLiteral(dict, i)
			if !ok {
				return ""
			}
			i = end
		case '<':
			if i+1 < len(dict) && dict[i+1] == '<' {
				i += 2
				continue
			}
			end, ok := scanPDFHexString(dict, i)
			if !ok {
				return ""
			}
			i = end
		case '/':
			nameEnd := scanPDFNameEnd(dict, i+1)
			if dict[i+1:nameEnd] == key {
				valueStart := skipPDFWhitespace(dict, nameEnd)
				valueEnd, ok := scanPDFValue(dict, valueStart)
				if !ok {
					return ""
				}
				return dict[valueStart:valueEnd]
			}
			i = nameEnd
		default:
			i++
		}
	}
	return ""
}

func decodePDFString(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "(") && strings.HasSuffix(raw, ")") && len(raw) >= 2 {
		return decodePDFBytes(decodePDFLiteral(raw[1 : len(raw)-1]))
	}
	if strings.HasPrefix(raw, "<") && strings.HasSuffix(raw, ">") {
		hex := compactPDFHex(raw[1 : len(raw)-1])
		if len(hex)%2 == 1 {
			hex += "0"
		}
		buf := make([]byte, 0, len(hex)/2)
		for i := 0; i+1 < len(hex); i += 2 {
			v, err := strconv.ParseUint(hex[i:i+2], 16, 8)
			if err != nil {
				return ""
			}
			buf = append(buf, byte(v))
		}
		return decodePDFBytes(buf)
	}
	if strings.HasPrefix(raw, "/") {
		return strings.TrimPrefix(raw, "/")
	}
	return raw
}

func scanPDFValue(s string, start int) (int, bool) {
	if start >= len(s) {
		return 0, false
	}
	switch s[start] {
	case '(':
		return scanPDFLiteral(s, start)
	case '<':
		if start+1 < len(s) && s[start+1] == '<' {
			end, ok := matchDictEnd(s, start)
			if !ok {
				return 0, false
			}
			return end + 2, true
		}
		return scanPDFHexString(s, start)
	case '/':
		return scanPDFNameEnd(s, start+1), true
	default:
		end := start
		for end < len(s) && !isPDFDelimiter(s[end]) {
			end++
		}
		return end, end > start
	}
}

func scanPDFLiteral(s string, start int) (int, bool) {
	if start >= len(s) || s[start] != '(' {
		return 0, false
	}
	depth := 1
	for i := start + 1; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i + 1, true
			}
		}
	}
	return 0, false
}

func scanPDFHexString(s string, start int) (int, bool) {
	if start >= len(s) || s[start] != '<' {
		return 0, false
	}
	for i := start + 1; i < len(s); i++ {
		if s[i] == '>' {
			return i + 1, true
		}
	}
	return 0, false
}

func scanPDFNameEnd(s string, start int) int {
	end := start
	for end < len(s) && !isPDFDelimiter(s[end]) {
		end++
	}
	return end
}

func skipPDFWhitespace(s string, start int) int {
	for start < len(s) && isPDFWhitespace(s[start]) {
		start++
	}
	return start
}

func isPDFDelimiter(c byte) bool {
	if isPDFWhitespace(c) {
		return true
	}
	switch c {
	case '(', ')', '<', '>', '[', ']', '{', '}', '/', '%':
		return true
	default:
		return false
	}
}

func isPDFWhitespace(c byte) bool {
	switch c {
	case 0x00, '\t', '\n', '\f', '\r', ' ':
		return true
	default:
		return false
	}
}

func decodePDFLiteral(s string) []byte {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			out = append(out, s[i])
			continue
		}
		i++
		if i >= len(s) {
			break
		}
		switch s[i] {
		case 'n':
			out = append(out, '\n')
		case 'r':
			out = append(out, '\r')
		case 't':
			out = append(out, '\t')
		case 'b':
			out = append(out, '\b')
		case 'f':
			out = append(out, '\f')
		case '(', ')', '\\':
			out = append(out, s[i])
		case '\r':
			if i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
		case '\n':
		default:
			if s[i] >= '0' && s[i] <= '7' {
				end := i + 1
				for end < len(s) && end < i+3 && s[end] >= '0' && s[end] <= '7' {
					end++
				}
				v, err := strconv.ParseUint(s[i:end], 8, 8)
				if err == nil {
					out = append(out, byte(v))
				}
				i = end - 1
				continue
			}
			out = append(out, s[i])
		}
	}
	return out
}

func compactPDFHex(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if !isPDFWhitespace(s[i]) {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func decodePDFBytes(b []byte) string {
	if len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF {
		return decodeUTF16BE(b[2:])
	}
	return string(b)
}

func decodeUTF16BE(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	units := make([]uint16, 0, len(b)/2)
	for i := 0; i+1 < len(b); i += 2 {
		units = append(units, uint16(b[i])<<8|uint16(b[i+1]))
	}
	return string(utf16.Decode(units))
}

func mergeMetadata(primary model.Metadata, fallback model.Metadata) model.Metadata {
	out := primary
	if out.Title == "" {
		out.Title = fallback.Title
	}
	if out.Author == "" {
		out.Author = fallback.Author
	}
	if out.Subject == "" {
		out.Subject = fallback.Subject
	}
	if out.Keywords == "" {
		out.Keywords = fallback.Keywords
	}
	if out.Creator == "" {
		out.Creator = fallback.Creator
	}
	if out.Producer == "" {
		out.Producer = fallback.Producer
	}
	if out.CreationDate == "" {
		out.CreationDate = fallback.CreationDate
	}
	if out.ModDate == "" {
		out.ModDate = fallback.ModDate
	}
	return out
}

func upsertNamedRef(dict, key string, ref objRef) string {
	replacement := fmt.Sprintf(`/%s %d %d R`, key, ref.Obj, ref.Gen)
	re := regexp.MustCompile(`/` + regexp.QuoteMeta(key) + `\s+\d+\s+\d+\s+R`)
	if re.MatchString(dict) {
		return re.ReplaceAllString(dict, replacement)
	}
	idx := strings.LastIndex(dict, ">>")
	if idx < 0 {
		return dict
	}
	insert := "\n" + replacement + "\n"
	return dict[:idx] + insert + dict[idx:]
}

func renderInfoObject(objNr int, m model.Metadata) []byte {
	type kv struct{ k, v string }
	entries := []kv{
		{"Title", m.Title},
		{"Author", m.Author},
		{"Subject", m.Subject},
		{"Keywords", m.Keywords},
		{"Creator", m.Creator},
		{"Producer", m.Producer},
		{"CreationDate", m.CreationDate},
		{"ModDate", m.ModDate},
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].k < entries[j].k })

	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d 0 obj\n<<", objNr))
	for _, e := range entries {
		if strings.TrimSpace(e.v) == "" {
			continue
		}
		b.WriteString(fmt.Sprintf("\n/%s %s", e.k, encodePDFString(e.v)))
	}
	b.WriteString("\n>>\nendobj\n")
	return []byte(b.String())
}

func renderMetadataObject(objNr int, packet []byte) []byte {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d 0 obj\n", objNr))
	b.WriteString(fmt.Sprintf("<< /Type /Metadata /Subtype /XML /Length %d >>\n", len(packet)))
	b.WriteString("stream\n")
	b.Write(packet)
	if !bytes.HasSuffix(packet, []byte("\n")) {
		b.WriteString("\n")
	}
	b.WriteString("endstream\nendobj\n")
	return []byte(b.String())
}

func renderCatalogObject(objNr int, dict string) []byte {
	return []byte(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objNr, strings.TrimSpace(dict)))
}

func renderXRef(startObj int, offsets []int) []byte {
	var b strings.Builder
	b.WriteString("xref\n")
	b.WriteString(fmt.Sprintf("%d %d\n", startObj, len(offsets)))
	for _, off := range offsets {
		b.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
	}
	return []byte(b.String())
}

func renderTrailer(size int, root objRef, info objRef, prev int) []byte {
	return []byte(fmt.Sprintf("trailer\n<< /Size %d /Root %d %d R /Info %d %d R /Prev %d >>\n", size, root.Obj, root.Gen, info.Obj, info.Gen, prev))
}

func escapePDFLiteral(s string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"(", "\\(",
		")", "\\)",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	return replacer.Replace(s)
}

func encodePDFString(s string) string {
	if needsUTF16BE(s) {
		return "<" + encodeUTF16BEHex(s) + ">"
	}
	return "(" + escapePDFLiteral(s) + ")"
}

func needsUTF16BE(s string) bool {
	for _, r := range s {
		if r > 0x7F {
			return true
		}
	}
	return false
}

func encodeUTF16BEHex(s string) string {
	units := utf16.Encode([]rune(s))
	var b strings.Builder
	b.Grow(4 + len(units)*4)
	b.WriteString("FEFF")
	for _, unit := range units {
		b.WriteString(fmt.Sprintf("%04X", unit))
	}
	return b.String()
}
