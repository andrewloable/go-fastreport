package script_test

import (
	"image/color"
	"testing"

	"github.com/andrewloable/go-fastreport/script"
	"github.com/andrewloable/go-fastreport/style"
)

type testShape struct {
	name    string
	visible bool
	fill    style.Fill
}

func (s *testShape) ScriptGetProperty(p string) interface{} {
	switch p {
	case "Visible":
		return s.visible
	case "Fill":
		return s.fill
	}
	return nil
}

func (s *testShape) ScriptSetProperty(p string, v interface{}) {
	switch p {
	case "Visible":
		if b, ok := v.(bool); ok {
			s.visible = b
		}
	case "Fill":
		if f, ok := v.(style.Fill); ok {
			s.fill = f
		}
	}
}

func (s *testShape) Name() string { return s.name }

const oitmScript = `
private void Cell4_BeforePrint(object sender, EventArgs e)
{
  decimal value = Cell4.Value == null ? 0 : (decimal)Cell4.Value;
  Shape1.Visible = true;
  Shape2.Visible = value >= 100;
  Shape3.Visible = value >= 3000;
  Color color = Color.Red;
  if (value >= 100)
    color = Color.Yellow;
  if (value >= 3000)
    color = Color.GreenYellow;
  Shape1.Fill = new SolidFill(color);
  Shape2.Fill = new SolidFill(color);
  Shape3.Fill = new SolidFill(color);
}
`

func TestParseScript_OitM(t *testing.T) {
	s, err := script.ParseScript(oitmScript)
	if err != nil {
		t.Fatalf("ParseScript error: %v", err)
	}
	handler, ok := s.Methods["Cell4_BeforePrint"]
	if !ok {
		t.Fatal("Cell4_BeforePrint not found")
	}

	tests := []struct {
		value     float64
		wantS1Vis bool
		wantS2Vis bool
		wantS3Vis bool
		wantFillR uint8
		wantFillG uint8
		wantFillB uint8
	}{
		{50, true, false, false, 255, 0, 0},    // Red
		{150, true, true, false, 255, 255, 0},  // Yellow
		{5000, true, true, true, 173, 255, 47}, // GreenYellow
	}

	for _, tt := range tests {
		s1 := &testShape{name: "Shape1"}
		s2 := &testShape{name: "Shape2"}
		s3 := &testShape{name: "Shape3"}

		ctx := &script.Context{
			SenderName:  "Cell4",
			SenderValue: tt.value,
			Objects: map[string]script.ContextObject{
				"Shape1": s1,
				"Shape2": s2,
				"Shape3": s3,
			},
			Vars: make(map[string]interface{}),
		}
		handler(ctx)

		if s1.visible != tt.wantS1Vis {
			t.Errorf("value=%.0f: Shape1.Visible=%v, want %v", tt.value, s1.visible, tt.wantS1Vis)
		}
		if s2.visible != tt.wantS2Vis {
			t.Errorf("value=%.0f: Shape2.Visible=%v, want %v", tt.value, s2.visible, tt.wantS2Vis)
		}
		if s3.visible != tt.wantS3Vis {
			t.Errorf("value=%.0f: Shape3.Visible=%v, want %v", tt.value, s3.visible, tt.wantS3Vis)
		}

		sf, ok := s1.fill.(*style.SolidFill)
		if !ok {
			t.Errorf("value=%.0f: Shape1.Fill is not *style.SolidFill, got %T", tt.value, s1.fill)
			continue
		}
		want := color.RGBA{R: tt.wantFillR, G: tt.wantFillG, B: tt.wantFillB, A: 255}
		if sf.Color != want {
			t.Errorf("value=%.0f: Shape1.Fill.Color=%v, want %v", tt.value, sf.Color, want)
		}
	}
}

const printMonthNamesScript = `
  public class ReportScript
  {
    private string[] monthNames = new string[] {
      "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec" };

    private void Cell4_BeforePrint(object sender, EventArgs e)
    {
      Cell4.Text = monthNames[(int)Cell4.Value - 1];
    }
  }
`

type testCell struct {
	text         string
	currentValue interface{}
}

func (c *testCell) ScriptGetProperty(name string) interface{} {
	switch name {
	case "Text":
		return c.text
	case "Value":
		return c.currentValue
	}
	return nil
}

func (c *testCell) ScriptSetProperty(name string, value interface{}) {
	if name == "Text" {
		if s, ok := value.(string); ok {
			c.text = s
		}
	}
}

func TestParseScript_PrintMonthNames(t *testing.T) {
	s, err := script.ParseScript(printMonthNamesScript)
	if err != nil {
		t.Fatalf("ParseScript error: %v", err)
	}
	handler, ok := s.Methods["Cell4_BeforePrint"]
	if !ok {
		t.Fatal("Cell4_BeforePrint not found")
	}

	tests := []struct {
		monthNum float64
		wantName string
	}{
		{1, "Jan"},
		{2, "Feb"},
		{3, "Mar"},
		{10, "Oct"},
		{11, "Nov"},
		{12, "Dec"},
	}

	for _, tt := range tests {
		cell := &testCell{currentValue: tt.monthNum}
		ctx := &script.Context{
			SenderName:  "Cell4",
			SenderValue: tt.monthNum,
			Objects: map[string]script.ContextObject{
				"Cell4": cell,
			},
			Vars: make(map[string]interface{}),
		}
		handler(ctx)
		if cell.text != tt.wantName {
			t.Errorf("monthNum=%.0f: got %q, want %q", tt.monthNum, cell.text, tt.wantName)
		}
	}
}
