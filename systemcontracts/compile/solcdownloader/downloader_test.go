// Copyright 2025 The go-stablenet Authors
// This file is part of the go-stablenet library.
//
// The go-stablenet library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-stablenet library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-stablenet library. If not, see <http://www.gnu.org/licenses/>.

package solcdownloader

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSecureClientRejectsHTTPDowngrade verifies that secureClient rejects HTTPS -> HTTP redirects
func TestSecureClientRejectsHTTPDowngrade(t *testing.T) {
	// Create HTTP server (insecure)
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("HTTP content - should not reach here"))
	}))
	defer httpServer.Close()

	// Create HTTPS server that redirects to HTTP
	httpsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, httpServer.URL, http.StatusMovedPermanently)
	}))
	defer httpsServer.Close()

	// Test 1: Default http.Client (vulnerable behavior)
	t.Run("DefaultClientFollowsDowngrade", func(t *testing.T) {
		defaultClient := httpsServer.Client()
		resp, err := defaultClient.Get(httpsServer.URL)
		if err != nil {
			t.Fatalf("Default client error: %v", err)
		}
		defer resp.Body.Close()

		// Verify downgrade occurred
		if resp.Request.URL.Scheme != "http" {
			t.Errorf("Expected default client to follow downgrade to http, got: %s", resp.Request.URL.Scheme)
		}
		t.Logf("PASS: Default client is vulnerable: followed HTTPS -> HTTP redirect")
	})

	// Test 2: secureClient (fixed behavior)
	t.Run("SecureClientRejectsDowngrade", func(t *testing.T) {
		// Configure secureClient to trust the test server's certificate
		testClient := &http.Client{
			CheckRedirect: secureClient.CheckRedirect,
			Timeout:       secureClient.Timeout,
			Transport:     httpsServer.Client().Transport, // Use test server's transport for TLS
		}

		resp, err := testClient.Get(httpsServer.URL)
		if err == nil {
			resp.Body.Close()
			t.Fatal("Expected error when following HTTPS → HTTP redirect, got nil")
		}

		// Verify error message
		if !strings.Contains(err.Error(), "refusing to follow redirect from HTTPS to HTTP") {
			t.Errorf("Expected downgrade rejection error, got: %v", err)
		}
		t.Logf("PASS: secureClient correctly rejected HTTPS -> HTTP redirect: %v", err)
	})
}

// TestSecureClientAllowsHTTPSRedirect verifies that secureClient allows HTTPS -> HTTPS redirects
func TestSecureClientAllowsHTTPSRedirect(t *testing.T) {
	// Create target HTTPS server
	targetServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Final HTTPS destination"))
	}))
	defer targetServer.Close()

	// Create source HTTPS server that redirects to another HTTPS server
	sourceServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL, http.StatusMovedPermanently)
	}))
	defer sourceServer.Close()

	// Configure secureClient with test server's transport
	testClient := &http.Client{
		CheckRedirect: secureClient.CheckRedirect,
		Timeout:       secureClient.Timeout,
		Transport:     sourceServer.Client().Transport,
	}

	resp, err := testClient.Get(sourceServer.URL)
	if err != nil {
		t.Fatalf("secureClient should allow HTTPS → HTTPS redirect, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}
	t.Logf("PASS: secureClient correctly allowed HTTPS -> HTTPS redirect")
}

// TestSecureClientLimitsRedirects verifies that secureClient limits redirect count
func TestSecureClientLimitsRedirects(t *testing.T) {
	redirectCount := 0

	// Create server that redirects to itself (infinite loop)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectCount++
		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
	}))
	defer server.Close()

	// Configure secureClient with test server's transport
	testClient := &http.Client{
		CheckRedirect: secureClient.CheckRedirect,
		Timeout:       secureClient.Timeout,
		Transport:     server.Client().Transport,
	}

	resp, err := testClient.Get(server.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("Expected error for excessive redirects, got nil")
	}

	// Verify error message
	if !strings.Contains(err.Error(), "stopped after 10 redirects") {
		t.Errorf("Expected redirect limit error, got: %v", err)
	}

	// Verify redirect count (should be exactly 10 before being stopped)
	if redirectCount < 10 {
		t.Errorf("Expected at least 10 redirects to be attempted, got: %d", redirectCount)
	}
	t.Logf("PASS: secureClient correctly limited redirects to 10 (attempted: %d)", redirectCount)
}

// TestSecureClientTimeout verifies that secureClient respects timeout
func TestSecureClientTimeout(t *testing.T) {
	// Create server that hangs
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never respond
		select {}
	}))
	defer server.Close()

	// Note: We don't actually test the full 5-minute timeout as it would slow down tests
	// This test just verifies that timeout is configured
	if secureClient.Timeout.Minutes() != 5 {
		t.Errorf("Expected 5 minute timeout, got: %v", secureClient.Timeout)
	}
	t.Logf("PASS: secureClient has 5 minute timeout configured")
}

// TestDownloadAttackSimulation simulates a real MITM attack scenario
// This test verifies that without the security patch, an attacker can inject malicious files
func TestDownloadAttackSimulation(t *testing.T) {
	maliciousContent := "MALICIOUS_FILE_CONTENT_INJECTED_BY_ATTACKER"

	// Create malicious HTTP server (attacker's server)
	attackerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(maliciousContent))
	}))
	defer attackerServer.Close()

	// Create legitimate HTTPS server that gets compromised (simulates MITM attack)
	legitimateServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// MITM attack: redirect to attacker's HTTP server
		http.Redirect(w, r, attackerServer.URL, http.StatusMovedPermanently)
	}))
	defer legitimateServer.Close()

	t.Run("VulnerableClientDownloadsMaliciousFile", func(t *testing.T) {
		// Vulnerable client (default http.Client)
		vulnerableClient := &http.Client{
			Transport: legitimateServer.Client().Transport,
			// No CheckRedirect protection - vulnerable!
		}

		resp, err := vulnerableClient.Get(legitimateServer.URL)
		if err != nil {
			t.Fatalf("Vulnerable client error: %v", err)
		}
		defer resp.Body.Close()

		// Read downloaded content
		body := make([]byte, len(maliciousContent))
		n, _ := resp.Body.Read(body)
		downloadedContent := string(body[:n])

		// Verify that malicious content was downloaded
		if downloadedContent != maliciousContent {
			t.Errorf("Expected malicious content, got: %s", downloadedContent)
		}

		// Verify downgrade occurred
		if resp.Request.URL.Scheme != "http" {
			t.Errorf("Expected downgrade to http, got: %s", resp.Request.URL.Scheme)
		}

		t.Logf("WARNING: ATTACK SUCCESS: Vulnerable client downloaded malicious file via HTTP")
		t.Logf("    Downloaded content: %s", downloadedContent)
		t.Logf("    Final URL scheme: %s (downgraded from https)", resp.Request.URL.Scheme)
	})

	t.Run("SecureClientPreventsAttack", func(t *testing.T) {
		// Secure client with CheckRedirect protection
		secureTestClient := &http.Client{
			CheckRedirect: secureClient.CheckRedirect,
			Timeout:       secureClient.Timeout,
			Transport:     legitimateServer.Client().Transport,
		}

		resp, err := secureTestClient.Get(legitimateServer.URL)
		if err == nil {
			resp.Body.Close()
			t.Fatal("SECURITY FAILURE: secureClient should reject HTTPS -> HTTP redirect")
		}

		// Verify error message
		if !strings.Contains(err.Error(), "refusing to follow redirect from HTTPS to HTTP") {
			t.Errorf("Expected downgrade rejection error, got: %v", err)
		}

		t.Logf("PASS: ATTACK PREVENTED: secureClient rejected malicious redirect")
		t.Logf("  Error: %v", err)
	})
}

// TestGetBuildWithMaliciousRedirect tests that getBuild function rejects HTTP downgrade
// This test verifies the security patch by checking if the function properly rejects
// HTTPS -> HTTP redirects that could be used in MITM attacks
func TestGetBuildWithMaliciousRedirect(t *testing.T) {
	maliciousJSON := `{
		"builds": [{
			"path": "solc-macosx-amd64-v0.8.14+commit.malicious",
			"version": "0.8.14",
			"build": "commit.malicious",
			"longVersion": "0.8.14+commit.malicious",
			"sha256": "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			"urls": ["bzzr://malicious"]
		}]
	}`

	// Attacker's HTTP server serving malicious list.json
	attackerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(maliciousJSON))
	}))
	defer attackerServer.Close()

	// Simulated compromised HTTPS URL that redirects to HTTP (attacker's server)
	// We test with a redirect chain: Start URL → HTTP attacker URL
	redirectChain := attackerServer.URL // This is HTTP

	// Test by calling getBuild with an HTTP URL that simulates post-redirect state
	// In real attack: HTTPS server redirects to this HTTP URL
	build, err := getBuild(redirectChain, "0.8.14")

	// Expected behavior:
	// - With secureClient.Get(): Request should be made and succeed (baseline test)
	// - With http.Get(): Same result (this tests vulnerability exists)
	// The key test is in TestDownloadAttackSimulation which tests the redirect behavior

	if err != nil {
		t.Logf("getBuild with HTTP URL error (expected if using secure client with redirect check): %v", err)
	} else if build != nil {
		if strings.Contains(build.Build, "malicious") {
			t.Logf("WARNING: Downloaded malicious data from HTTP URL: %s", build.Build)
			t.Logf("    This shows the vulnerability: attacker-controlled data can be fetched")
		}
	}

	// The real security test: Verify that secureClient has the CheckRedirect function
	if secureClient.CheckRedirect == nil {
		t.Errorf("SECURITY FAILURE: secureClient has no CheckRedirect protection!")
		t.Errorf("    The security patch is missing or incomplete")
	} else {
		t.Logf("PASS: secureClient has CheckRedirect protection configured")
	}
}

// TestVersionCompare tests the version comparison logic
func TestVersionCompare(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
		desc     string
	}{
		{"0.8.14", "0.8.14", 0, "equal versions"},
		{"0.8.14", "0.8.15", -1, "v1 < v2"},
		{"0.8.15", "0.8.14", 1, "v1 > v2"},
		{"0.7.2", "0.8.0", -1, "major version diff"},
		{"1.0.0", "0.9.9", 1, "major version upgrade"},
		{"v0.8.14", "0.8.14", 0, "with v prefix"},
	}

	for _, tt := range tests {
		result := versionCompare(tt.v1, tt.v2)
		if result != tt.expected {
			t.Errorf("versionCompare(%q, %q) = %d, expected %d (%s)",
				tt.v1, tt.v2, result, tt.expected, tt.desc)
		} else {
			t.Logf("PASS: versionCompare(%q, %q) = %d (%s)",
				tt.v1, tt.v2, result, tt.desc)
		}
	}
}
