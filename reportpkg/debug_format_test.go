package reportpkg_test

import (
	"fmt"
	"testing"
	"github.com/andrewloable/go-fastreport/format"
	"github.com/andrewloable/go-fastreport/object"
	"github.com/andrewloable/go-fastreport/report"
)

func TestDebugFormat(t *testing.T) {
	r := loadFRXSmoke(t, "Master-Detail.frx")
	type hasObjects interface {
		Objects() *report.ObjectCollection
	}
	for _, pg := range r.Pages() {
		for _, b := range pg.AllBands() {
			if h, ok := b.(hasObjects); ok {
				objs := h.Objects()
				for i := 0; i < objs.Len(); i++ {
					if to, ok := objs.Get(i).(*object.TextObject); ok {
						f := to.Format()
						fmt.Printf("TextObject %q format=%T\n", to.Name(), f)
						if cf, ok := f.(*format.CurrencyFormat); ok {
							fmt.Printf("  UseLocaleSettings=%v\n", cf.UseLocaleSettings)
						}
					}
				}
			}
		}
	}
}
