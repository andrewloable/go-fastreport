package utils

import (
	"strings"
	"testing"
)

func TestStripRTF_PlainText(t *testing.T) {
	result := StripRTF("Hello World")
	if result != "Hello World" {
		t.Errorf("StripRTF plain text = %q, want %q", result, "Hello World")
	}
}

func TestStripRTF_Bold(t *testing.T) {
	rtf := `{\rtf1\ansi{\b bold text}}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "bold text") {
		t.Errorf("StripRTF should preserve text content, got %q", result)
	}
}

func TestStripRTF_Par(t *testing.T) {
	rtf := `{\rtf1 line1\par line2}`
	result := StripRTF(rtf)
	if !strings.Contains(result, "\n") {
		t.Errorf("StripRTF should convert \\par to newline, got %q", result)
	}
}

func TestRTFToHTML_PlainText(t *testing.T) {
	result := RTFToHTML("Hello World")
	if !strings.Contains(result, "Hello World") {
		t.Errorf("RTFToHTML plain = %q, want to contain 'Hello World'", result)
	}
}

func TestRTFToHTML_Bold(t *testing.T) {
	rtf := `{\rtf1\ansi {\b Bold Text}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<b>") || !strings.Contains(result, "</b>") {
		t.Errorf("RTFToHTML should emit <b> tags for \\b, got: %q", result)
	}
	if !strings.Contains(result, "Bold Text") {
		t.Errorf("RTFToHTML should preserve text content, got: %q", result)
	}
}

func TestRTFToHTML_Italic(t *testing.T) {
	rtf := `{\rtf1\ansi {\i Italic Text}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<i>") || !strings.Contains(result, "</i>") {
		t.Errorf("RTFToHTML should emit <i> tags for \\i, got: %q", result)
	}
}

func TestRTFToHTML_Underline(t *testing.T) {
	rtf := `{\rtf1\ansi {\ul Underlined}}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<u>") || !strings.Contains(result, "</u>") {
		t.Errorf("RTFToHTML should emit <u> tags for \\ul, got: %q", result)
	}
}

func TestRTFToHTML_ParagraphBreak(t *testing.T) {
	rtf := `{\rtf1 Para 1\par Para 2}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "<br>") {
		t.Errorf("RTFToHTML should emit <br> for \\par, got: %q", result)
	}
	if !strings.Contains(result, "Para 1") || !strings.Contains(result, "Para 2") {
		t.Errorf("RTFToHTML should preserve both paragraphs, got: %q", result)
	}
}

func TestRTFToHTML_HTMLEscaping(t *testing.T) {
	rtf := `{\rtf1 a < b & c > d}`
	result := RTFToHTML(rtf)
	if strings.Contains(result, "<b") && !strings.Contains(result, "&lt;") {
		t.Errorf("RTFToHTML should HTML-escape < in plain text, got: %q", result)
	}
}

func TestRTFToHTML_SpecialChars(t *testing.T) {
	rtf := `{\rtf1\ansi \endash \emdash}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "&ndash;") {
		t.Errorf("RTFToHTML should emit &ndash; for \\endash, got: %q", result)
	}
	if !strings.Contains(result, "&mdash;") {
		t.Errorf("RTFToHTML should emit &mdash; for \\emdash, got: %q", result)
	}
}

func TestRTFToHTML_MixedFormatting(t *testing.T) {
	rtf := `{\rtf1 Normal {\b Bold {\i Both}} Normal}`
	result := RTFToHTML(rtf)
	if !strings.Contains(result, "Normal") {
		t.Errorf("RTFToHTML should preserve normal text, got: %q", result)
	}
	if !strings.Contains(result, "<b>") {
		t.Errorf("RTFToHTML should emit bold, got: %q", result)
	}
}
