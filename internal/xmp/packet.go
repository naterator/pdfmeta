package xmp

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"

	"pdfmeta/internal/model"
)

const (
	xpacketBegin = `<?xpacket begin="\uFEFF" id="W5M0MpCehiHzreSzNTczkc9d"?>`
	xpacketEnd   = `<?xpacket end="w"?>`
	xmpOpenTag   = "<x:xmpmeta"
	xmpCloseTag  = "</x:xmpmeta>"
)

// Marshal converts canonical metadata into an XMP packet.
func Marshal(m model.Metadata) ([]byte, error) {
	var b strings.Builder
	b.WriteString(xpacketBegin + "\n")
	b.WriteString(`<x:xmpmeta xmlns:x="adobe:ns:meta/" x:xmptk="pdfmeta">` + "\n")
	b.WriteString(`<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">` + "\n")
	b.WriteString(`<rdf:Description rdf:about=""`)
	b.WriteString(` xmlns:dc="http://purl.org/dc/elements/1.1/"`)
	b.WriteString(` xmlns:pdf="http://ns.adobe.com/pdf/1.3/"`)
	b.WriteString(` xmlns:xmp="http://ns.adobe.com/xap/1.0/">` + "\n")

	writeLangAlt(&b, "dc:title", m.Title)
	writeSeq(&b, "dc:creator", m.Author)
	writeLangAlt(&b, "dc:description", m.Subject)
	writeValue(&b, "pdf:Keywords", m.Keywords)
	writeValue(&b, "xmp:CreatorTool", m.Creator)
	writeValue(&b, "pdf:Producer", m.Producer)
	writeValue(&b, "xmp:CreateDate", m.CreationDate)
	writeValue(&b, "xmp:ModifyDate", m.ModDate)

	b.WriteString(`</rdf:Description>` + "\n")
	b.WriteString(`</rdf:RDF>` + "\n")
	b.WriteString(`</x:xmpmeta>` + "\n")
	b.WriteString(xpacketEnd + "\n")
	return []byte(b.String()), nil
}

func writeLangAlt(b *strings.Builder, key, value string) {
	if value == "" {
		return
	}
	b.WriteString("<" + key + "><rdf:Alt><rdf:li xml:lang=\"x-default\">")
	xmlEscape(b, value)
	b.WriteString("</rdf:li></rdf:Alt></" + key + ">\n")
}

func writeSeq(b *strings.Builder, key, value string) {
	if value == "" {
		return
	}
	b.WriteString("<" + key + "><rdf:Seq><rdf:li>")
	xmlEscape(b, value)
	b.WriteString("</rdf:li></rdf:Seq></" + key + ">\n")
}

func writeValue(b *strings.Builder, key, value string) {
	if value == "" {
		return
	}
	b.WriteString("<" + key + ">")
	xmlEscape(b, value)
	b.WriteString("</" + key + ">\n")
}

func xmlEscape(b *strings.Builder, s string) {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	b.WriteString(buf.String())
}

// Unmarshal parses an XMP packet and maps known fields into canonical metadata.
func Unmarshal(packet []byte) (model.Metadata, error) {
	dec := xml.NewDecoder(bytes.NewReader(packet))
	var (
		stack  []string
		meta   model.Metadata
		sawXMP bool
	)

	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return model.Metadata{}, fmt.Errorf("decode xml: %w", err)
		}

		switch t := tok.(type) {
		case xml.StartElement:
			stack = append(stack, t.Name.Local)
			if t.Name.Local == "xmpmeta" {
				sawXMP = true
			}
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			value := strings.TrimSpace(string(t))
			if value == "" {
				continue
			}
			switch {
			case hasSuffix(stack, "title", "Alt", "li"):
				if meta.Title == "" {
					meta.Title = value
				}
			case hasSuffix(stack, "creator", "Seq", "li"):
				if meta.Author == "" {
					meta.Author = value
				}
			case hasSuffix(stack, "description", "Alt", "li"):
				if meta.Subject == "" {
					meta.Subject = value
				}
			case hasSuffix(stack, "Keywords"):
				meta.Keywords = value
			case hasSuffix(stack, "CreatorTool"):
				meta.Creator = value
			case hasSuffix(stack, "Producer"):
				meta.Producer = value
			case hasSuffix(stack, "CreateDate"):
				meta.CreationDate = value
			case hasSuffix(stack, "ModifyDate"):
				meta.ModDate = value
			}
		}
	}

	if !sawXMP {
		return model.Metadata{}, errors.New("xmp packet not found")
	}
	return meta, nil
}

func hasSuffix(stack []string, want ...string) bool {
	if len(stack) < len(want) {
		return false
	}
	start := len(stack) - len(want)
	for i := range want {
		if stack[start+i] != want[i] {
			return false
		}
	}
	return true
}

// Extract returns the first XMP packet found in PDF bytes.
func Extract(pdfBytes []byte) ([]byte, bool) {
	start := bytes.Index(pdfBytes, []byte(xmpOpenTag))
	if start < 0 {
		return nil, false
	}
	end := bytes.Index(pdfBytes[start:], []byte(xmpCloseTag))
	if end < 0 {
		return nil, false
	}
	end += start + len(xmpCloseTag)
	return append([]byte(nil), pdfBytes[start:end]...), true
}

// Upsert inserts a packet before %%EOF or replaces the existing xmpmeta section.
func Upsert(pdfBytes []byte, metadata model.Metadata) ([]byte, error) {
	packet, err := Marshal(metadata)
	if err != nil {
		return nil, err
	}

	start := bytes.Index(pdfBytes, []byte(xmpOpenTag))
	if start >= 0 {
		end := bytes.Index(pdfBytes[start:], []byte(xmpCloseTag))
		if end >= 0 {
			end += start + len(xmpCloseTag)
			out := make([]byte, 0, len(pdfBytes)-(end-start)+len(packet))
			out = append(out, pdfBytes[:start]...)
			out = append(out, packet...)
			out = append(out, pdfBytes[end:]...)
			return out, nil
		}
	}

	eof := bytes.LastIndex(pdfBytes, []byte("%%EOF"))
	if eof < 0 {
		return append(append([]byte(nil), pdfBytes...), packet...), nil
	}
	out := make([]byte, 0, len(pdfBytes)+len(packet))
	out = append(out, pdfBytes[:eof]...)
	out = append(out, packet...)
	out = append(out, pdfBytes[eof:]...)
	return out, nil
}
