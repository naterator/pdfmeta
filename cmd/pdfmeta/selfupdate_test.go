package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

type stubReleaseUpdater struct {
	runCalls       int
	currentVersion string
	err            error
}

func (s *stubReleaseUpdater) Run(_ context.Context, currentVersion string, stdout io.Writer) error {
	s.runCalls++
	s.currentVersion = currentVersion
	if stdout != nil {
		_, _ = io.WriteString(stdout, "stub updater invoked\n")
	}
	return s.err
}

func withReleaseUpdaterStub(t *testing.T, updater releaseUpdater) {
	t.Helper()
	previous := makeReleaseUpdater
	makeReleaseUpdater = func() releaseUpdater { return updater }
	t.Cleanup(func() {
		makeReleaseUpdater = previous
	})
}

func TestRunSelfupdateCommandRunsUpdater(t *testing.T) {
	stub := &stubReleaseUpdater{}
	withReleaseUpdaterStub(t, stub)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"selfupdate"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if stub.runCalls != 1 {
		t.Fatalf("stub updater runCalls = %d, want 1", stub.runCalls)
	}
	if stub.currentVersion != appVersion {
		t.Fatalf("stub updater currentVersion = %q, want %q", stub.currentVersion, appVersion)
	}
	if !strings.Contains(stdout.String(), "stub updater invoked") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRunSelfupdateCommandRejectsExtraArgs(t *testing.T) {
	stub := &stubReleaseUpdater{}
	withReleaseUpdaterStub(t, stub)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"selfupdate", "unexpected-arg"}, &stdout, &stderr)
	if exitCode == 0 {
		t.Fatalf("expected non-zero exit code for extra args")
	}
	if stub.runCalls != 0 {
		t.Fatalf("stub updater runCalls = %d, want 0", stub.runCalls)
	}
}

func TestRunSelfupdateReportsUpdaterError(t *testing.T) {
	stub := &stubReleaseUpdater{err: io.EOF}
	withReleaseUpdaterStub(t, stub)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"selfupdate"}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if stub.runCalls != 1 {
		t.Fatalf("stub updater runCalls = %d, want 1", stub.runCalls)
	}
	if !strings.Contains(stderr.String(), "selfupdate failed: EOF") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestRunVersionCommand(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"version"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr=%q", exitCode, stderr.String())
	}
	if got := stdout.String(); got != appVersion+"\n" {
		t.Fatalf("stdout = %q, want %q", got, appVersion+"\\n")
	}
}

func TestGitHubReleaseUpdaterReplacesExecutable(t *testing.T) {
	t.Parallel()
	exePath := filepath.Join(t.TempDir(), appName)
	if err := os.WriteFile(exePath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	binaryCandidates, checksumCandidates := releaseAssetCandidates(runtime.GOOS, runtime.GOARCH)
	binaryName := binaryCandidates[0]
	checksumName := checksumCandidates[0]
	binaryBody := []byte("new-binary-content")
	sum := sha256.Sum256(binaryBody)
	checksumBody := hex.EncodeToString(sum[:]) + "  " + binaryName + "\n"

	var metadataCalls int
	var binaryCalls int
	var checksumCalls int
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/naterator/pdfmeta/releases/latest":
			metadataCalls++
			if got := r.Header.Get("User-Agent"); got != appName+"/"+appVersion {
				t.Fatalf("User-Agent = %q", got)
			}
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v9.9.9",
				Assets: []githubReleaseAsset{
					{Name: binaryName, BrowserDownloadURL: server.URL + "/download/" + binaryName},
					{Name: checksumName, BrowserDownloadURL: server.URL + "/download/" + checksumName},
				},
			})
		case "/download/" + binaryName:
			binaryCalls++
			_, _ = w.Write(binaryBody)
		case "/download/" + checksumName:
			checksumCalls++
			_, _ = io.WriteString(w, checksumBody)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	updater := &githubReleaseUpdater{
		client:           server.Client(),
		latestReleaseURL: server.URL + "/repos/naterator/pdfmeta/releases/latest",
		executablePath: func() (string, error) {
			return exePath, nil
		},
		goos:   runtime.GOOS,
		goarch: runtime.GOARCH,
	}

	var stdout bytes.Buffer
	if err := updater.Run(context.Background(), "1.0.0", &stdout); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(got) != string(binaryBody) {
		t.Fatalf("updated executable content = %q, want %q", string(got), string(binaryBody))
	}
	info, err := os.Stat(exePath)
	if err != nil {
		t.Fatalf("Stat returned error: %v", err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("updated executable mode = %#o, want 0755", info.Mode().Perm())
	}
	if metadataCalls != 1 || binaryCalls != 1 || checksumCalls != 1 {
		t.Fatalf("calls = metadata:%d binary:%d checksum:%d", metadataCalls, binaryCalls, checksumCalls)
	}
	if !strings.Contains(stdout.String(), "Updating "+appName+" from 1.0.0 to v9.9.9") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Updated "+appName+" to v9.9.9") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestGitHubReleaseUpdaterSkipsCurrentVersion(t *testing.T) {
	t.Parallel()
	exePath := filepath.Join(t.TempDir(), appName)
	if err := os.WriteFile(exePath, []byte("existing-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var metadataCalls int
	var downloadCalls int
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/naterator/pdfmeta/releases/latest":
			metadataCalls++
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v1.0.2",
				Assets:  []githubReleaseAsset{},
			})
		default:
			downloadCalls++
			t.Fatalf("unexpected download path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	updater := &githubReleaseUpdater{
		client:           server.Client(),
		latestReleaseURL: server.URL + "/repos/naterator/pdfmeta/releases/latest",
		executablePath: func() (string, error) {
			return exePath, nil
		},
		goos:   runtime.GOOS,
		goarch: runtime.GOARCH,
	}

	var stdout bytes.Buffer
	if err := updater.Run(context.Background(), appVersion, &stdout); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if string(got) != "existing-binary" {
		t.Fatalf("executable content = %q, want existing-binary", string(got))
	}
	if metadataCalls != 1 || downloadCalls != 0 {
		t.Fatalf("calls = metadata:%d download:%d", metadataCalls, downloadCalls)
	}
	if !strings.Contains(stdout.String(), "already up to date") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestGitHubReleaseUpdaterRejectsChecksumMismatch(t *testing.T) {
	t.Parallel()
	exePath := filepath.Join(t.TempDir(), appName)
	if err := os.WriteFile(exePath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	binaryCandidates, checksumCandidates := releaseAssetCandidates(runtime.GOOS, runtime.GOARCH)
	binaryName := binaryCandidates[0]
	checksumName := checksumCandidates[0]

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/naterator/pdfmeta/releases/latest":
			_ = json.NewEncoder(w).Encode(githubRelease{
				TagName: "v9.9.9",
				Assets: []githubReleaseAsset{
					{Name: binaryName, BrowserDownloadURL: server.URL + "/download/" + binaryName},
					{Name: checksumName, BrowserDownloadURL: server.URL + "/download/" + checksumName},
				},
			})
		case "/download/" + binaryName:
			_, _ = io.WriteString(w, "tampered-binary")
		case "/download/" + checksumName:
			_, _ = io.WriteString(w, strings.Repeat("a", 64)+"  "+binaryName+"\n")
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	updater := &githubReleaseUpdater{
		client:           server.Client(),
		latestReleaseURL: server.URL + "/repos/naterator/pdfmeta/releases/latest",
		executablePath: func() (string, error) {
			return exePath, nil
		},
		goos:   runtime.GOOS,
		goarch: runtime.GOARCH,
	}

	err := updater.Run(context.Background(), "1.0.0", io.Discard)
	if err == nil {
		t.Fatal("Run unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("Run error = %v", err)
	}

	got, readErr := os.ReadFile(exePath)
	if readErr != nil {
		t.Fatalf("ReadFile returned error: %v", readErr)
	}
	if string(got) != "old-binary" {
		t.Fatalf("executable content = %q, want old-binary", string(got))
	}
}

func TestNormalizeSemver(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{input: "v1.0.0", want: "v1.0.0"},
		{input: "1.0.0", want: "v1.0.0"},
		{input: "v0.0.1", want: "v0.0.1"},
		{input: "v10.20.30", want: "v10.20.30"},
		{input: " v1.2.3 ", want: "v1.2.3"},
		{input: "", wantErr: true},
		{input: "v1.0", wantErr: true},
		{input: "v1.0.0.0", wantErr: true},
		{input: "v1.0.a", wantErr: true},
		{input: "v1..0", wantErr: true},
		{input: "v01.0.0", wantErr: true},
		{input: "v1.00.0", wantErr: true},
		{input: "v1.0.00", wantErr: true},
		{input: "v1.0.0-beta", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := normalizeSemver(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeSemver(%q) = %q, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeSemver(%q) error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeSemver(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeSemverLoose(t *testing.T) {
	t.Parallel()
	if got, ok := normalizeSemverLoose("v1.2.3"); !ok || got != "v1.2.3" {
		t.Fatalf("normalizeSemverLoose(v1.2.3) = %q, %v", got, ok)
	}
	if _, ok := normalizeSemverLoose("garbage"); ok {
		t.Fatalf("normalizeSemverLoose(garbage) unexpectedly succeeded")
	}
	if _, ok := normalizeSemverLoose(""); ok {
		t.Fatalf("normalizeSemverLoose(\"\") unexpectedly succeeded")
	}
}

func TestCompareSemver(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		want int
	}{
		{a: "v1.0.0", b: "v1.0.0", want: 0},
		{a: "v2.0.0", b: "v1.0.0", want: 1},
		{a: "v1.0.0", b: "v2.0.0", want: -1},
		{a: "v1.1.0", b: "v1.0.0", want: 1},
		{a: "v1.0.1", b: "v1.0.0", want: 1},
		{a: "v1.0.0", b: "v1.0.1", want: -1},
		{a: "v1.0.10", b: "v1.0.9", want: 1},
		{a: "v1.0.9", b: "v1.0.10", want: -1},
		{a: "v10.0.0", b: "v9.0.0", want: 1},
		{a: "v9.0.0", b: "v10.0.0", want: -1},
		{a: "v0.0.0", b: "v0.0.0", want: 0},
		{a: "v100.200.300", b: "v100.200.300", want: 0},
		{a: "v1.0.2", b: "v1.0.1", want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			t.Parallel()
			if got := compareSemver(tt.a, tt.b); got != tt.want {
				t.Fatalf("compareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestAssetsForRuntimeRejectsWindows(t *testing.T) {
	t.Parallel()
	release := githubRelease{
		TagName: "v1.0.0",
		Assets: []githubReleaseAsset{
			{Name: "pdfmeta-windows-amd64.exe", BrowserDownloadURL: "https://example.com/bin"},
			{Name: "pdfmeta-windows-amd64.exe.sha256", BrowserDownloadURL: "https://example.com/sum"},
		},
	}
	_, _, err := release.assetsForRuntime("windows", "amd64")
	if err == nil {
		t.Fatal("assetsForRuntime(windows) unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "not supported on windows") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseSHA256File(t *testing.T) {
	t.Parallel()
	hash := strings.Repeat("ab", 32)

	t.Run("named match", func(t *testing.T) {
		t.Parallel()
		got, err := parseSHA256File(hash+"  mybin\n", "mybin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != hash {
			t.Fatalf("got %q, want %q", got, hash)
		}
	})

	t.Run("single unnamed", func(t *testing.T) {
		t.Parallel()
		got, err := parseSHA256File(hash+"\n", "anything")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != hash {
			t.Fatalf("got %q, want %q", got, hash)
		}
	})

	t.Run("multiple unnamed", func(t *testing.T) {
		t.Parallel()
		_, err := parseSHA256File(hash+"\n"+strings.Repeat("cd", 32)+"\n", "anything")
		if err == nil {
			t.Fatal("expected error for multiple unnamed digests")
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()
		_, err := parseSHA256File(hash+"  otherbin\n", "mybin")
		if err == nil {
			t.Fatal("expected error for missing asset")
		}
	})

	t.Run("key=value format", func(t *testing.T) {
		t.Parallel()
		got, err := parseSHA256File("SHA256="+hash+"\n", "mybin")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != hash {
			t.Fatalf("got %q, want %q", got, hash)
		}
	})
}
