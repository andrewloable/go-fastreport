package object

import (
	"image/color"

	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/report"
	"github.com/andrewloable/go-fastreport/style"
	"github.com/andrewloable/go-fastreport/utils"
)

// Duplicates specifies how duplicate values are displayed.
type Duplicates int

const (
	// DuplicatesShow allows the text object to show duplicate values.
	DuplicatesShow Duplicates = iota
	// DuplicatesHide hides text objects with duplicate values.
	DuplicatesHide
	// DuplicatesClear shows the text object but with no text when the value is duplicate.
	DuplicatesClear
	// DuplicatesMerge merges multiple text objects with the same value into one.
	DuplicatesMerge
)

// ProcessAt specifies when the report engine processes a text object.
type ProcessAt int

const (
	// ProcessAtDefault processes just-in-time (default).
	ProcessAtDefault ProcessAt = iota
	// ProcessAtReportFinished processes when the entire report is finished.
	ProcessAtReportFinished
	// ProcessAtReportPageFinished processes when the entire report page is finished.
	ProcessAtReportPageFinished
	// ProcessAtPageFinished processes when any report page is finished.
	ProcessAtPageFinished
	// ProcessAtColumnFinished processes when the column is finished.
	ProcessAtColumnFinished
	// ProcessAtDataFinished processes when the data block is finished.
	ProcessAtDataFinished
	// ProcessAtGroupFinished processes when the group is finished.
	ProcessAtGroupFinished
	// ProcessAtCustom processes manually via Engine.ProcessObject.
	ProcessAtCustom
)

// formatProcessAt converts ProcessAt to its FRX string name.
func formatProcessAt(p ProcessAt) string {
	switch p {
	case ProcessAtReportFinished:
		return "ReportFinished"
	case ProcessAtReportPageFinished:
		return "ReportPageFinished"
	case ProcessAtPageFinished:
		return "PageFinished"
	case ProcessAtColumnFinished:
		return "ColumnFinished"
	case ProcessAtDataFinished:
		return "DataFinished"
	case ProcessAtGroupFinished:
		return "GroupFinished"
	case ProcessAtCustom:
		return "Custom"
	default:
		return "Default"
	}
}

// parseProcessAt converts an FRX string to ProcessAt (handles both names and ints).
func parseProcessAt(s string) ProcessAt {
	switch s {
	case "ReportFinished", "1":
		return ProcessAtReportFinished
	case "ReportPageFinished", "2":
		return ProcessAtReportPageFinished
	case "PageFinished", "3":
		return ProcessAtPageFinished
	case "ColumnFinished", "4":
		return ProcessAtColumnFinished
	case "DataFinished", "5":
		return ProcessAtDataFinished
	case "GroupFinished", "6":
		return ProcessAtGroupFinished
	case "Custom", "7":
		return ProcessAtCustom
	default:
		return ProcessAtDefault
	}
}

// formatDuplicates converts Duplicates to its FRX string name.
func formatDuplicates(d Duplicates) string {
	switch d {
	case DuplicatesHide:
		return "Hide"
	case DuplicatesClear:
		return "Clear"
	case DuplicatesMerge:
		return "Merge"
	default:
		return "Show"
	}
}

// parseDuplicates converts an FRX string to Duplicates (handles both names and ints).
func parseDuplicates(s string) Duplicates {
	switch s {
	case "Hide", "1":
		return DuplicatesHide
	case "Clear", "2":
		return DuplicatesClear
	case "Merge", "3":
		return DuplicatesMerge
	default:
		return DuplicatesShow
	}
}

// Padding holds interior spacing (left, top, right, bottom) in pixels.
type Padding struct {
	Left, Top, Right, Bottom float32
}

// TextObjectBase is the base for text objects (TextObject).
// It is the Go equivalent of FastReport.TextObjectBase.
type TextObjectBase struct {
	report.BreakableComponent

	text             string
	allowExpressions bool   // default true
	brackets         string // default "[,]"
	padding          Padding
	hideZeros        bool
	hideValue        string
	nullValue        string
	processAt        ProcessAt
	duplicates       Duplicates
	editable         bool

	// format is the value formatter applied to evaluated expression results.
	// nil means GeneralFormat (default fmt.Sprint behavior).
	format format.Format
}

// NewTextObjectBase creates a TextObjectBase with defaults.
// C# default padding is (2, 0, 2, 0) — left=2, top=0, right=2, bottom=0.
func NewTextObjectBase() *TextObjectBase {
	t := &TextObjectBase{
		BreakableComponent: *report.NewBreakableComponent(),
		allowExpressions:   true,
		brackets:           "[,]",
		padding:            Padding{Left: 2, Top: 0, Right: 2, Bottom: 0},
	}
	return t
}

// --- Text ---

// Text returns the object's text (may contain expressions like "[Date]").
func (t *TextObjectBase) Text() string { return t.text }

// SetText sets the text.
func (t *TextObjectBase) SetText(s string) { t.text = s }

// --- Expression handling ---

// AllowExpressions returns whether the text may contain expressions.
func (t *TextObjectBase) AllowExpressions() bool { return t.allowExpressions }

// SetAllowExpressions sets the allow-expressions flag.
func (t *TextObjectBase) SetAllowExpressions(v bool) { t.allowExpressions = v }

// Brackets returns the expression delimiter symbols (e.g. "[,]").
func (t *TextObjectBase) Brackets() string { return t.brackets }

// SetBrackets sets the expression delimiters.
func (t *TextObjectBase) SetBrackets(s string) { t.brackets = s }

// --- Padding ---

// Padding returns interior padding.
func (t *TextObjectBase) Padding() Padding { return t.padding }

// SetPadding sets interior padding.
func (t *TextObjectBase) SetPadding(p Padding) { t.padding = p }

// --- Zero/null hiding ---

// HideZeros returns whether zero values are hidden.
func (t *TextObjectBase) HideZeros() bool { return t.hideZeros }

// SetHideZeros sets the hide-zeros flag.
func (t *TextObjectBase) SetHideZeros(v bool) { t.hideZeros = v }

// HideValue returns the specific value string that is hidden.
func (t *TextObjectBase) HideValue() string { return t.hideValue }

// SetHideValue sets the specific value to hide.
func (t *TextObjectBase) SetHideValue(s string) { t.hideValue = s }

// NullValue returns the text shown when a data value is null/nil.
func (t *TextObjectBase) NullValue() string { return t.nullValue }

// SetNullValue sets the null-replacement text.
func (t *TextObjectBase) SetNullValue(s string) { t.nullValue = s }

// --- Process timing ---

// ProcessAt returns when this object is processed by the engine.
func (t *TextObjectBase) ProcessAt() ProcessAt { return t.processAt }

// SetProcessAt sets the process timing.
func (t *TextObjectBase) SetProcessAt(p ProcessAt) { t.processAt = p }

// --- Duplicates ---

// Duplicates returns how duplicate values are handled.
func (t *TextObjectBase) Duplicates() Duplicates { return t.duplicates }

// SetDuplicates sets the duplicate-handling mode.
func (t *TextObjectBase) SetDuplicates(d Duplicates) { t.duplicates = d }

// --- Editable ---

// Editable returns whether the text can be edited in the viewer.
func (t *TextObjectBase) Editable() bool { return t.editable }

// SetEditable sets the editable flag.
func (t *TextObjectBase) SetEditable(v bool) { t.editable = v }

// --- Format ---

// Format returns the value formatter for this text object.
// Returns nil when no explicit format has been set (GeneralFormat).
func (t *TextObjectBase) Format() format.Format { return t.format }

// SetFormat sets the value formatter.
func (t *TextObjectBase) SetFormat(f format.Format) { t.format = f }

// --- Serialization ---

// Serialize writes TextObjectBase properties that differ from defaults.
func (t *TextObjectBase) Serialize(w report.Writer) error {
	if err := t.BreakableComponent.Serialize(w); err != nil {
		return err
	}
	if t.text != "" {
		w.WriteStr("Text", t.text)
	}
	if !t.allowExpressions {
		w.WriteBool("AllowExpressions", false)
	}
	if t.brackets != "[,]" {
		w.WriteStr("Brackets", t.brackets)
	}
	if t.padding != (Padding{}) {
		w.WriteStr("Padding", paddingToStr(t.padding))
	}
	if t.hideZeros {
		w.WriteBool("HideZeros", true)
	}
	if t.hideValue != "" {
		w.WriteStr("HideValue", t.hideValue)
	}
	if t.nullValue != "" {
		w.WriteStr("NullValue", t.nullValue)
	}
	if t.processAt != ProcessAtDefault {
		w.WriteStr("ProcessAt", formatProcessAt(t.processAt))
	}
	if t.duplicates != DuplicatesShow {
		w.WriteStr("Duplicates", formatDuplicates(t.duplicates))
	}
	if t.editable {
		w.WriteBool("Editable", true)
	}
	if t.format != nil {
		serializeTextFormat(w, t.format)
	}
	return nil
}

// Deserialize reads TextObjectBase properties.
func (t *TextObjectBase) Deserialize(r report.Reader) error {
	if err := t.BreakableComponent.Deserialize(r); err != nil {
		return err
	}
	t.text = r.ReadStr("Text", "")
	t.allowExpressions = r.ReadBool("AllowExpressions", true)
	t.brackets = r.ReadStr("Brackets", "[,]")
	if s := r.ReadStr("Padding", ""); s != "" {
		t.padding = strToPadding(s)
	}
	t.hideZeros = r.ReadBool("HideZeros", false)
	t.hideValue = r.ReadStr("HideValue", "")
	t.nullValue = r.ReadStr("NullValue", "")
	t.processAt = parseProcessAt(r.ReadStr("ProcessAt", "Default"))
	t.duplicates = parseDuplicates(r.ReadStr("Duplicates", "Show"))
	t.editable = r.ReadBool("Editable", false)
	if ft := r.ReadStr("Format", ""); ft != "" {
		t.format = deserializeTextFormat(ft, r)
	}
	return nil
}

// paddingToStr serializes a Padding as "L,T,R,B".
func paddingToStr(p Padding) string {
	return report.FormatFloat(p.Left) + "," + report.FormatFloat(p.Top) + "," +
		report.FormatFloat(p.Right) + "," + report.FormatFloat(p.Bottom)
}

// strToPadding parses "L,T,R,B".
func strToPadding(s string) Padding {
	parts := report.SplitComma(s)
	if len(parts) != 4 {
		return Padding{}
	}
	return Padding{
		Left:   report.ParseFloat(parts[0]),
		Top:    report.ParseFloat(parts[1]),
		Right:  report.ParseFloat(parts[2]),
		Bottom: report.ParseFloat(parts[3]),
	}
}

// -----------------------------------------------------------------------
// HorzAlign, VertAlign, AutoShrinkMode, LineSpacingType
// -----------------------------------------------------------------------

// HorzAlign specifies horizontal text alignment.
type HorzAlign int

const (
	HorzAlignLeft    HorzAlign = iota
	HorzAlignCenter
	HorzAlignRight
	HorzAlignJustify
)

// VertAlign specifies vertical text alignment.
type VertAlign int

const (
	VertAlignTop    VertAlign = iota
	VertAlignCenter
	VertAlignBottom
)

// ParseHorzAlign parses a horizontal alignment value from an FRX attribute.
// FRX files may store alignment as a string name ("Left", "Center", "Right",
// "Justify") or as a numeric string ("0", "1", "2", "3").
func ParseHorzAlign(s string) HorzAlign {
	switch s {
	case "Center", "1":
		return HorzAlignCenter
	case "Right", "2":
		return HorzAlignRight
	case "Justify", "3":
		return HorzAlignJustify
	default:
		return HorzAlignLeft
	}
}

// formatHorzAlign converts HorzAlign to its FRX string name.
func formatHorzAlign(h HorzAlign) string {
	switch h {
	case HorzAlignCenter:
		return "Center"
	case HorzAlignRight:
		return "Right"
	case HorzAlignJustify:
		return "Justify"
	default:
		return "Left"
	}
}

// ParseVertAlign parses a vertical alignment value from an FRX attribute.
// FRX files may store alignment as a string name ("Top", "Center", "Bottom")
// or as a numeric string ("0", "1", "2").
func ParseVertAlign(s string) VertAlign {
	switch s {
	case "Center", "1":
		return VertAlignCenter
	case "Bottom", "2":
		return VertAlignBottom
	default:
		return VertAlignTop
	}
}

// formatVertAlign converts VertAlign to its FRX string name.
func formatVertAlign(v VertAlign) string {
	switch v {
	case VertAlignCenter:
		return "Center"
	case VertAlignBottom:
		return "Bottom"
	default:
		return "Top"
	}
}

// AutoShrinkMode controls the AutoShrink feature.
type AutoShrinkMode int

const (
	AutoShrinkNone     AutoShrinkMode = iota
	AutoShrinkFontSize
	AutoShrinkFontWidth
	AutoShrinkFitText
)

// formatAutoShrinkMode converts AutoShrinkMode to its FRX string name.
func formatAutoShrinkMode(m AutoShrinkMode) string {
	switch m {
	case AutoShrinkFontSize:
		return "FontSize"
	case AutoShrinkFontWidth:
		return "FontWidth"
	case AutoShrinkFitText:
		return "FitText"
	default:
		return "None"
	}
}

// parseAutoShrinkMode converts an FRX string to AutoShrinkMode (handles both names and ints).
func parseAutoShrinkMode(s string) AutoShrinkMode {
	switch s {
	case "FontSize", "1":
		return AutoShrinkFontSize
	case "FontWidth", "2":
		return AutoShrinkFontWidth
	case "FitText", "3":
		return AutoShrinkFitText
	default:
		return AutoShrinkNone
	}
}

// LineSpacingType controls line-spacing calculation.
type LineSpacingType int

const (
	LineSpacingSingle   LineSpacingType = iota
	LineSpacingAtLeast
	LineSpacingExactly
	LineSpacingMultiple
)

// ParagraphFormat holds paragraph-level formatting.
type ParagraphFormat struct {
	FirstLineIndent    float32
	LineSpacing        float32
	LineSpacingType    LineSpacingType
	SkipFirstLineIndent bool
}

// MergeMode flags for TextObject.
type MergeMode int

const (
	MergeModeNone       MergeMode = 0
	MergeModeHorizontal MergeMode = 1
	MergeModeVertical   MergeMode = 2
)

// formatMergeMode converts MergeMode to its FRX string name.
func formatMergeMode(m MergeMode) string {
	switch m {
	case MergeModeHorizontal:
		return "Horizontal"
	case MergeModeVertical:
		return "Vertical"
	default:
		return "None"
	}
}

// parseMergeMode converts an FRX string to MergeMode (handles both names and ints).
func parseMergeMode(s string) MergeMode {
	switch s {
	case "Horizontal", "1":
		return MergeModeHorizontal
	case "Vertical", "2":
		return MergeModeVertical
	default:
		return MergeModeNone
	}
}

// TextRenderType selects the rendering engine.
type TextRenderType int

const (
	TextRenderTypeDefault      TextRenderType = iota
	TextRenderTypeHtmlTags
	TextRenderTypeHtmlParagraph
	TextRenderTypeInline
)

// formatTextRenderType converts TextRenderType to its FRX string name.
func formatTextRenderType(t TextRenderType) string {
	switch t {
	case TextRenderTypeHtmlTags:
		return "HtmlTags"
	case TextRenderTypeHtmlParagraph:
		return "HtmlParagraph"
	case TextRenderTypeInline:
		return "Inline"
	default:
		return "Default"
	}
}

// parseTextRenderType converts an FRX string to TextRenderType (handles both names and ints).
func parseTextRenderType(s string) TextRenderType {
	switch s {
	case "HtmlTags", "1":
		return TextRenderTypeHtmlTags
	case "HtmlParagraph", "2":
		return TextRenderTypeHtmlParagraph
	case "Inline", "3":
		return TextRenderTypeInline
	default:
		return TextRenderTypeDefault
	}
}

// -----------------------------------------------------------------------
// TextObject
// -----------------------------------------------------------------------

// TextObject represents a text element that may display one or more lines.
// It is the Go equivalent of FastReport.TextObject.
type TextObject struct {
	TextObjectBase

	horzAlign         HorzAlign
	vertAlign         VertAlign
	angle             int
	rightToLeft       bool
	wordWrap          bool // default true
	underlines        bool
	font              style.Font
	// textColor is the foreground text color (equivalent to FastReport TextFill as SolidFill).
	// Default is opaque black {0, 0, 0, 255}.
	textColor         color.RGBA
	fontWidthRatio    float32 // default 1.0
	firstTabOffset    float32
	tabWidth          float32
	clip              bool // default true
	wysiwyg           bool
	lineHeight        float32
	forceJustify      bool
	textRenderType    TextRenderType
	autoShrink        AutoShrinkMode
	autoShrinkMinSize float32
	paragraphOffset   float32
	paragraphFormat   ParagraphFormat
	mergeMode         MergeMode
	autoWidth         bool

	// textOutline defines an optional stroke drawn around each character.
	textOutline style.TextOutline

	// highlights holds the conditional-formatting rules for this text object.
	// The engine evaluates each condition in order and applies the first match.
	highlights []style.HighlightCondition

	// formats is the multi-format collection for multi-expression text objects.
	// When non-nil, the formats here supersede the single TextObjectBase.format.
	formats *format.Collection
}

// NewTextObject creates a TextObject with defaults.
func NewTextObject() *TextObject {
	return &TextObject{
		TextObjectBase: *NewTextObjectBase(),
		wordWrap:       true,
		fontWidthRatio: 1.0,
		clip:           true,
		font:           style.DefaultFont(),
		textColor:      color.RGBA{A: 255}, // opaque black
	}
}

// --- Alignment ---

// HorzAlign returns the horizontal text alignment.
func (t *TextObject) HorzAlign() HorzAlign { return t.horzAlign }

// SetHorzAlign sets the horizontal alignment.
func (t *TextObject) SetHorzAlign(a HorzAlign) { t.horzAlign = a }

// VertAlign returns the vertical text alignment.
func (t *TextObject) VertAlign() VertAlign { return t.vertAlign }

// SetVertAlign sets the vertical alignment.
func (t *TextObject) SetVertAlign(a VertAlign) { t.vertAlign = a }

// --- Angle ---

// Angle returns the text rotation in degrees.
func (t *TextObject) Angle() int { return t.angle }

// SetAngle sets the text rotation.
func (t *TextObject) SetAngle(a int) { t.angle = a }

// --- Text direction ---

// RightToLeft returns whether text flows right-to-left.
func (t *TextObject) RightToLeft() bool { return t.rightToLeft }

// SetRightToLeft sets the right-to-left flag.
func (t *TextObject) SetRightToLeft(v bool) { t.rightToLeft = v }

// WordWrap returns whether text wraps to multiple lines.
func (t *TextObject) WordWrap() bool { return t.wordWrap }

// SetWordWrap sets word-wrap.
func (t *TextObject) SetWordWrap(v bool) { t.wordWrap = v }

// Underlines returns whether each text line has an underline drawn below it.
func (t *TextObject) Underlines() bool { return t.underlines }

// SetUnderlines sets the underlines flag.
func (t *TextObject) SetUnderlines(v bool) { t.underlines = v }

// --- Font ---

// Font returns the text font.
func (t *TextObject) Font() style.Font { return t.font }

// SetFont sets the text font.
func (t *TextObject) SetFont(f style.Font) { t.font = f }

// TextColor returns the foreground text color.
// The default is opaque black {0, 0, 0, 255}.
func (t *TextObject) TextColor() color.RGBA { return t.textColor }

// SetTextColor sets the foreground text color.
func (t *TextObject) SetTextColor(c color.RGBA) { t.textColor = c }

// ApplyStyle overrides the base implementation to also propagate the font and
// text-fill colour from the style entry. Both the modern Apply* flags and the
// legacy *Changed flags are honoured.
func (t *TextObject) ApplyStyle(entry *style.StyleEntry) {
	t.ReportComponentBase.ApplyStyle(entry)
	if entry == nil {
		return
	}
	if entry.ApplyFont || entry.FontChanged {
		t.font = entry.Font
	}
	if entry.ApplyTextFill || entry.TextColorChanged {
		t.textColor = entry.TextColor
	}
}

// TextOutline returns the text stroke/outline settings.
func (t *TextObject) TextOutline() style.TextOutline { return t.textOutline }

// SetTextOutline sets the text stroke/outline settings.
func (t *TextObject) SetTextOutline(v style.TextOutline) { t.textOutline = v }

// FontWidthRatio returns the horizontal font scaling ratio (default 1.0).
func (t *TextObject) FontWidthRatio() float32 { return t.fontWidthRatio }

// SetFontWidthRatio sets the font width ratio.
func (t *TextObject) SetFontWidthRatio(r float32) { t.fontWidthRatio = r }

// --- Tabs ---

// FirstTabOffset returns the offset of the first tab stop in pixels.
func (t *TextObject) FirstTabOffset() float32 { return t.firstTabOffset }

// SetFirstTabOffset sets the first tab offset.
func (t *TextObject) SetFirstTabOffset(v float32) { t.firstTabOffset = v }

// TabWidth returns the tab stop interval in pixels.
func (t *TextObject) TabWidth() float32 { return t.tabWidth }

// SetTabWidth sets the tab width.
func (t *TextObject) SetTabWidth(v float32) { t.tabWidth = v }

// --- Rendering options ---

// Clip returns whether text is clipped to the object bounds.
func (t *TextObject) Clip() bool { return t.clip }

// SetClip sets the clip flag.
func (t *TextObject) SetClip(v bool) { t.clip = v }

// Wysiwyg returns whether the object is rendered in WYSIWYG mode.
func (t *TextObject) Wysiwyg() bool { return t.wysiwyg }

// SetWysiwyg sets the WYSIWYG flag.
func (t *TextObject) SetWysiwyg(v bool) { t.wysiwyg = v }

// LineHeight returns the fixed line height in pixels (0 = auto).
func (t *TextObject) LineHeight() float32 { return t.lineHeight }

// SetLineHeight sets the line height.
func (t *TextObject) SetLineHeight(v float32) { t.lineHeight = v }

// ForceJustify returns whether the last line is also justified.
func (t *TextObject) ForceJustify() bool { return t.forceJustify }

// SetForceJustify sets the force-justify flag.
func (t *TextObject) SetForceJustify(v bool) { t.forceJustify = v }

// TextRenderType returns the rendering engine selection.
func (t *TextObject) TextRenderType() TextRenderType { return t.textRenderType }

// SetTextRenderType sets the render type.
func (t *TextObject) SetTextRenderType(r TextRenderType) { t.textRenderType = r }

// AutoShrink returns the auto-shrink mode.
func (t *TextObject) AutoShrink() AutoShrinkMode { return t.autoShrink }

// SetAutoShrink sets the auto-shrink mode.
func (t *TextObject) SetAutoShrink(m AutoShrinkMode) { t.autoShrink = m }

// AutoShrinkMinSize returns the minimum font size for auto-shrink.
func (t *TextObject) AutoShrinkMinSize() float32 { return t.autoShrinkMinSize }

// SetAutoShrinkMinSize sets the minimum size for auto-shrink.
func (t *TextObject) SetAutoShrinkMinSize(v float32) { t.autoShrinkMinSize = v }

// ParagraphOffset returns the indentation of the second and subsequent lines.
func (t *TextObject) ParagraphOffset() float32 { return t.paragraphOffset }

// SetParagraphOffset sets the paragraph offset.
func (t *TextObject) SetParagraphOffset(v float32) { t.paragraphOffset = v }

// ParagraphFormat returns paragraph-level formatting.
func (t *TextObject) ParagraphFormat() ParagraphFormat { return t.paragraphFormat }

// SetParagraphFormat sets the paragraph format.
func (t *TextObject) SetParagraphFormat(pf ParagraphFormat) { t.paragraphFormat = pf }

// Formats returns the multi-format collection. Returns nil when no collection
// has been set (single-format mode via TextObjectBase.Format() is used instead).
func (t *TextObject) Formats() *format.Collection { return t.formats }

// SetFormats replaces the multi-format collection.
func (t *TextObject) SetFormats(fc *format.Collection) { t.formats = fc }

// MergeMode returns the merge-mode flags.
func (t *TextObject) MergeMode() MergeMode { return t.mergeMode }

// SetMergeMode sets the merge mode.
func (t *TextObject) SetMergeMode(m MergeMode) { t.mergeMode = m }

// AutoWidth returns whether the object width grows to fit text.
func (t *TextObject) AutoWidth() bool { return t.autoWidth }

// SetAutoWidth sets auto-width.
func (t *TextObject) SetAutoWidth(v bool) { t.autoWidth = v }

// --- Highlight conditions ---

// Highlights returns the conditional-formatting rules for this object.
func (t *TextObject) Highlights() []style.HighlightCondition { return t.highlights }

// SetHighlights replaces the highlights slice.
func (t *TextObject) SetHighlights(h []style.HighlightCondition) { t.highlights = h }

// AddHighlight appends a highlight condition.
func (t *TextObject) AddHighlight(c style.HighlightCondition) {
	t.highlights = append(t.highlights, c)
}

// DeserializeChild handles <Highlight> and <Formats> child elements from FRX.
// It satisfies report.ChildDeserializer so that reportpkg.deserializeChildren
// can delegate unknown child elements to the TextObject itself.
func (t *TextObject) DeserializeChild(childType string, r report.Reader) bool {
	if childType == "Formats" {
		// Multi-format case: populate the FormatCollection.
		t.formats = format.NewCollection()
		for {
			ct, ok := r.NextChild()
			if !ok {
				break
			}
			f := deserializeFormatFromChild(ct, r)
			if f != nil {
				t.formats.Add(f)
				// Keep the single format field in sync with the primary.
				if t.format == nil {
					t.format = f
				}
			}
			// drain any children of the format element
			for {
				_, ok2 := r.NextChild()
				if !ok2 {
					break
				}
				if r.FinishChild() != nil {
					break
				}
			}
			if r.FinishChild() != nil {
				break
			}
		}
		return true
	}
	if childType != "Highlight" {
		return false
	}
	// <Highlight> contains <Condition …/> child elements.
	for {
		condType, ok := r.NextChild()
		if !ok {
			break
		}
		if condType == "Condition" {
			c := style.NewHighlightCondition()
			c.Expression = r.ReadStr("Expression", "")
			c.Visible = r.ReadBool("Visible", true)
			c.ApplyBorder = r.ReadBool("ApplyBorder", false)
			c.ApplyFill = r.ReadBool("ApplyFill", false)
			c.ApplyFont = r.ReadBool("ApplyFont", false)
			c.ApplyTextFill = r.ReadBool("ApplyTextFill", true)
			if cs := r.ReadStr("Fill.Color", ""); cs != "" {
				if col, err := utils.ParseColor(cs); err == nil {
					c.FillColor = col
				}
			}
			if cs := r.ReadStr("TextFill.Color", ""); cs != "" {
				if col, err := utils.ParseColor(cs); err == nil {
					c.TextFillColor = col
				}
			}
			if fs := r.ReadStr("Font", ""); fs != "" {
				c.Font = style.FontFromStr(fs)
			}
			t.highlights = append(t.highlights, c)
		}
		if r.FinishChild() != nil { break }
	}
	return true
}

// --- Serialization ---

// Serialize writes TextObject properties that differ from defaults.
func (t *TextObject) Serialize(w report.Writer) error {
	if err := t.TextObjectBase.Serialize(w); err != nil {
		return err
	}
	if t.horzAlign != HorzAlignLeft {
		w.WriteStr("HorzAlign", formatHorzAlign(t.horzAlign))
	}
	if t.vertAlign != VertAlignTop {
		w.WriteStr("VertAlign", formatVertAlign(t.vertAlign))
	}
	if t.angle != 0 {
		w.WriteInt("Angle", t.angle)
	}
	if t.rightToLeft {
		w.WriteBool("RightToLeft", true)
	}
	if !t.wordWrap {
		w.WriteBool("WordWrap", false)
	}
	if t.underlines {
		w.WriteBool("Underlines", true)
	}
	if !style.FontEqual(t.font, style.DefaultFont()) {
		w.WriteStr("Font", style.FontToStr(t.font))
	}
	if t.fontWidthRatio != 1.0 {
		w.WriteFloat("FontWidthRatio", t.fontWidthRatio)
	}
	if t.firstTabOffset != 0 {
		w.WriteFloat("FirstTabOffset", t.firstTabOffset)
	}
	if t.tabWidth != 0 {
		w.WriteFloat("TabWidth", t.tabWidth)
	}
	if !t.clip {
		w.WriteBool("Clip", false)
	}
	if t.wysiwyg {
		w.WriteBool("Wysiwyg", true)
	}
	if t.lineHeight != 0 {
		w.WriteFloat("LineHeight", t.lineHeight)
	}
	if t.forceJustify {
		w.WriteBool("ForceJustify", true)
	}
	if t.textRenderType != TextRenderTypeDefault {
		w.WriteStr("TextRenderType", formatTextRenderType(t.textRenderType))
	}
	if t.autoShrink != AutoShrinkNone {
		w.WriteStr("AutoShrink", formatAutoShrinkMode(t.autoShrink))
	}
	if t.autoShrinkMinSize != 0 {
		w.WriteFloat("AutoShrinkMinSize", t.autoShrinkMinSize)
	}
	if t.paragraphOffset != 0 {
		w.WriteFloat("ParagraphOffset", t.paragraphOffset)
	}
	if t.mergeMode != MergeModeNone {
		w.WriteStr("MergeMode", formatMergeMode(t.mergeMode))
	}
	if t.autoWidth {
		w.WriteBool("AutoWidth", true)
	}
	// TextOutline
	if t.textOutline.Enabled {
		w.WriteBool("TextOutline.Enabled", true)
		w.WriteStr("TextOutline.Color", utils.FormatColor(t.textOutline.Color))
		if t.textOutline.Width != 1 {
			w.WriteFloat("TextOutline.Width", t.textOutline.Width)
		}
		if t.textOutline.DashStyle != 0 {
			w.WriteInt("TextOutline.DashStyle", t.textOutline.DashStyle)
		}
		if t.textOutline.DrawBehind {
			w.WriteBool("TextOutline.DrawBehind", true)
		}
	}
	if len(t.highlights) > 0 {
		coll := &conditionCollection{items: t.highlights}
		if err := w.WriteObjectNamed("Highlight", coll); err != nil {
			return err
		}
	}
	return nil
}

// Deserialize reads TextObject properties.
func (t *TextObject) Deserialize(r report.Reader) error {
	if err := t.TextObjectBase.Deserialize(r); err != nil {
		return err
	}
	t.horzAlign = ParseHorzAlign(r.ReadStr("HorzAlign", "Left"))
	t.vertAlign = ParseVertAlign(r.ReadStr("VertAlign", "Top"))
	t.angle = r.ReadInt("Angle", 0)
	t.rightToLeft = r.ReadBool("RightToLeft", false)
	t.wordWrap = r.ReadBool("WordWrap", true)
	t.underlines = r.ReadBool("Underlines", false)
	if s := r.ReadStr("Font", ""); s != "" {
		t.font = style.FontFromStr(s)
	}
	t.fontWidthRatio = r.ReadFloat("FontWidthRatio", 1.0)
	t.firstTabOffset = r.ReadFloat("FirstTabOffset", 0)
	t.tabWidth = r.ReadFloat("TabWidth", 0)
	t.clip = r.ReadBool("Clip", true)
	t.wysiwyg = r.ReadBool("Wysiwyg", false)
	t.lineHeight = r.ReadFloat("LineHeight", 0)
	t.forceJustify = r.ReadBool("ForceJustify", false)
	t.textRenderType = parseTextRenderType(r.ReadStr("TextRenderType", "Default"))
	t.autoShrink = parseAutoShrinkMode(r.ReadStr("AutoShrink", "None"))
	t.autoShrinkMinSize = r.ReadFloat("AutoShrinkMinSize", 0)
	t.paragraphOffset = r.ReadFloat("ParagraphOffset", 0)
	t.mergeMode = parseMergeMode(r.ReadStr("MergeMode", "None"))
	t.autoWidth = r.ReadBool("AutoWidth", false)
	// TextFill.Color — foreground text color (FastReport uses TextFill as a SolidFill).
	// This attribute is stored directly on the TextObject element, e.g. TextFill.Color="Brown".
	if cs := r.ReadStr("TextFill.Color", ""); cs != "" {
		if c, err := utils.ParseColor(cs); err == nil {
			t.textColor = c
		}
	}
	// TextOutline
	t.textOutline.Enabled = r.ReadBool("TextOutline.Enabled", false)
	if cs := r.ReadStr("TextOutline.Color", ""); cs != "" {
		if c, err := utils.ParseColor(cs); err == nil {
			t.textOutline.Color = c
		}
	} else {
		t.textOutline.Color = style.DefaultTextOutline().Color
	}
	t.textOutline.Width = r.ReadFloat("TextOutline.Width", 1)
	t.textOutline.DashStyle = r.ReadInt("TextOutline.DashStyle", 0)
	t.textOutline.DrawBehind = r.ReadBool("TextOutline.DrawBehind", false)
	return nil
}

// conditionCollection is an internal helper that implements report.Serializable
// for the HighlightCondition slice, matching the C# ConditionCollection which
// implements IFRSerializable with ItemName="Highlight".
type conditionCollection struct {
	items []style.HighlightCondition
}

func (c *conditionCollection) Serialize(w report.Writer) error {
	for _, cond := range c.items {
		hc := &highlightConditionSerializable{c: cond}
		if err := w.WriteObjectNamed("Condition", hc); err != nil {
			return err
		}
	}
	return nil
}

func (c *conditionCollection) Deserialize(r report.Reader) error {
	return nil
}

// highlightConditionSerializable wraps a HighlightCondition as report.Serializable.
type highlightConditionSerializable struct {
	c style.HighlightCondition
}

func (h *highlightConditionSerializable) Serialize(w report.Writer) error {
	def := style.NewHighlightCondition()
	if h.c.Expression != def.Expression {
		w.WriteStr("Expression", h.c.Expression)
	}
	if h.c.Visible != def.Visible {
		w.WriteBool("Visible", h.c.Visible)
	}
	if h.c.ApplyBorder != def.ApplyBorder {
		w.WriteBool("ApplyBorder", h.c.ApplyBorder)
	}
	if h.c.ApplyFill != def.ApplyFill {
		w.WriteBool("ApplyFill", h.c.ApplyFill)
	}
	if h.c.ApplyFont != def.ApplyFont {
		w.WriteBool("ApplyFont", h.c.ApplyFont)
	}
	if h.c.ApplyTextFill != def.ApplyTextFill {
		w.WriteBool("ApplyTextFill", h.c.ApplyTextFill)
	}
	if h.c.ApplyFill && h.c.FillColor != (color.RGBA{}) {
		w.WriteStr("Fill.Color", utils.FormatColor(h.c.FillColor))
	}
	if h.c.ApplyTextFill {
		w.WriteStr("TextFill.Color", utils.FormatColor(h.c.TextFillColor))
	}
	if h.c.ApplyFont {
		w.WriteStr("Font", style.FontToStr(h.c.Font))
	}
	return nil
}

func (h *highlightConditionSerializable) Deserialize(r report.Reader) error {
	return nil
}
