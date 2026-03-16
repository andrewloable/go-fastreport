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
//   gauge.go:315 LinearGauge.Serialize — 85.7% coverage
//     Line 317: `return err` after `g.GaugeObject.Serialize(w)`
//     GaugeObject.Serialize → ReportComponentBase.Serialize → ComponentBase.Serialize
//     → BaseObject.Serialize — all return nil unconditionally.
//
//   gauge.go:329 LinearGauge.Deserialize — 80.0% coverage
//     Line 331: `return err` after `g.GaugeObject.Deserialize(r)`
//     Same base-chain analysis.
//
//   gauge.go:373 RadialGauge.Serialize — 85.7% coverage
//     Line 375: `return err` after `g.GaugeObject.Serialize(w)`
//
//   gauge.go:387 RadialGauge.Deserialize — 80.0% coverage
//     Line 389: `return err` after `g.GaugeObject.Deserialize(r)`
//
//   gauge.go:450 SimpleGauge.Serialize — 94.1% coverage
//     Line 452: `return err` after `g.GaugeObject.Serialize(w)`
//
//   gauge.go:481 SimpleGauge.Deserialize — 90.0% coverage
//     Line 483: `return err` after `g.GaugeObject.Deserialize(r)`
//
//   gauge.go:523 SimpleProgressGauge.Serialize — 80.0% coverage
//     Line 525: `return err` after `g.GaugeObject.Serialize(w)`
//
//   gauge.go:534 SimpleProgressGauge.Deserialize — 75.0% coverage
//     Line 536: `return err` after `g.GaugeObject.Deserialize(r)`
//
// Additionally:
//   gauge.go:171 GaugeObject.Serialize — 97.3% coverage
//     Line 173: `return err` after `g.ReportComponentBase.Serialize(w)` — unreachable.
//
//   gauge.go:234 GaugeObject.Deserialize — 96.0% coverage
//     Line 236: `return err` after `g.ReportComponentBase.Deserialize(r)` — unreachable.
//
// The writer methods WriteStr, WriteInt, WriteBool, WriteFloat do not return
// errors (their signatures return void), so the entire base Serialize chain
// has no failure mode. Coverage for these functions is bounded by the number
// of these defensive `return err` statements.
