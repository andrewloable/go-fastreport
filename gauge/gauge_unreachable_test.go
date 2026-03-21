package gauge

// gauge_unreachable_test.go — documents structurally unreachable error branches
// in gauge.go that cannot be covered by tests.
//
// UNREACHABLE ANALYSIS
// ====================
//
// The following `return err` branches are structurally unreachable because the
// upstream Serialize/Deserialize chain (BaseObject → ComponentBase →
// ReportComponentBase → GaugeObject) always returns nil:
//
//   LinearGauge.Serialize — 85.7% coverage
//     `return err` after `g.GaugeObject.Serialize(w)`
//     GaugeObject.Serialize → ReportComponentBase.Serialize → ComponentBase.Serialize
//     → BaseObject.Serialize — all return nil unconditionally.
//
//   LinearGauge.Deserialize — 80.0% coverage
//     `return err` after `g.GaugeObject.Deserialize(r)` — same base-chain analysis.
//
//   RadialGauge.Serialize — 93.3% coverage (extra fields added in porting)
//     `return err` after `g.GaugeObject.Serialize(w)`
//
//   RadialGauge.Deserialize — 88.9% coverage
//     `return err` after `g.GaugeObject.Deserialize(r)`
//
//   SimpleGauge.Serialize — 94.1% coverage
//     `return err` after `g.GaugeObject.Serialize(w)`
//
//   SimpleGauge.Deserialize — 90.0% coverage
//     `return err` after `g.GaugeObject.Deserialize(r)`
//
//   SimpleProgressGauge.Serialize — 80.0% coverage
//     `return err` after `g.GaugeObject.Serialize(w)`
//
//   SimpleProgressGauge.Deserialize — 75.0% coverage
//     `return err` after `g.GaugeObject.Deserialize(r)`
//
// Additionally:
//   GaugeObject.Serialize — 97.3% coverage
//     `return err` after `g.ReportComponentBase.Serialize(w)` — unreachable.
//
//   GaugeObject.Deserialize — 96.0% coverage
//     `return err` after `g.ReportComponentBase.Deserialize(r)` — unreachable.
//
// The writer methods WriteStr, WriteInt, WriteBool, WriteFloat do not return
// errors (their signatures return void), so the entire base Serialize chain
// has no failure mode. Coverage for these functions is bounded by the number
// of these defensive `return err` statements.
