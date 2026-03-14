package object

import (
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
}

// NewTextObjectBase creates a TextObjectBase with defaults.
func NewTextObjectBase() *TextObjectBase {
	t := &TextObjectBase{
		BreakableComponent: *report.NewBreakableComponent(),
		allowExpressions:   true,
		brackets:           "[,]",
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
		w.WriteInt("ProcessAt", int(t.processAt))
	}
	if t.duplicates != DuplicatesShow {
		w.WriteInt("Duplicates", int(t.duplicates))
	}
	if t.editable {
		w.WriteBool("Editable", true)
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
	t.processAt = ProcessAt(r.ReadInt("ProcessAt", 0))
	t.duplicates = Duplicates(r.ReadInt("Duplicates", 0))
	t.editable = r.ReadBool("Editable", false)
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

// AutoShrinkMode controls the AutoShrink feature.
type AutoShrinkMode int

const (
	AutoShrinkNone     AutoShrinkMode = iota
	AutoShrinkFontSize
	AutoShrinkFontWidth
	AutoShrinkFitText
)

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

// TextRenderType selects the rendering engine.
type TextRenderType int

const (
	TextRenderTypeDefault      TextRenderType = iota
	TextRenderTypeHtmlTags
	TextRenderTypeHtmlParagraph
	TextRenderTypeInline
)

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

	// highlights holds the conditional-formatting rules for this text object.
	// The engine evaluates each condition in order and applies the first match.
	highlights []style.HighlightCondition
}

// NewTextObject creates a TextObject with defaults.
func NewTextObject() *TextObject {
	return &TextObject{
		TextObjectBase: *NewTextObjectBase(),
		wordWrap:       true,
		fontWidthRatio: 1.0,
		clip:           true,
		font:           style.DefaultFont(),
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

// AddHighlight appends a highlight condition.
func (t *TextObject) AddHighlight(c style.HighlightCondition) {
	t.highlights = append(t.highlights, c)
}

// DeserializeChild handles the <Highlight> child element from FRX.
// It satisfies report.ChildDeserializer so that reportpkg.deserializeChildren
// can delegate unknown child elements to the TextObject itself.
func (t *TextObject) DeserializeChild(childType string, r report.Reader) bool {
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
		_ = r.FinishChild()
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
		w.WriteInt("HorzAlign", int(t.horzAlign))
	}
	if t.vertAlign != VertAlignTop {
		w.WriteInt("VertAlign", int(t.vertAlign))
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
		w.WriteInt("TextRenderType", int(t.textRenderType))
	}
	if t.autoShrink != AutoShrinkNone {
		w.WriteInt("AutoShrink", int(t.autoShrink))
	}
	if t.autoShrinkMinSize != 0 {
		w.WriteFloat("AutoShrinkMinSize", t.autoShrinkMinSize)
	}
	if t.paragraphOffset != 0 {
		w.WriteFloat("ParagraphOffset", t.paragraphOffset)
	}
	if t.mergeMode != MergeModeNone {
		w.WriteInt("MergeMode", int(t.mergeMode))
	}
	if t.autoWidth {
		w.WriteBool("AutoWidth", true)
	}
	return nil
}

// Deserialize reads TextObject properties.
func (t *TextObject) Deserialize(r report.Reader) error {
	if err := t.TextObjectBase.Deserialize(r); err != nil {
		return err
	}
	t.horzAlign = HorzAlign(r.ReadInt("HorzAlign", 0))
	t.vertAlign = VertAlign(r.ReadInt("VertAlign", 0))
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
	t.textRenderType = TextRenderType(r.ReadInt("TextRenderType", 0))
	t.autoShrink = AutoShrinkMode(r.ReadInt("AutoShrink", 0))
	t.autoShrinkMinSize = r.ReadFloat("AutoShrinkMinSize", 0)
	t.paragraphOffset = r.ReadFloat("ParagraphOffset", 0)
	t.mergeMode = MergeMode(r.ReadInt("MergeMode", 0))
	t.autoWidth = r.ReadBool("AutoWidth", false)
	return nil
}
