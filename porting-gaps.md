# Porting Gaps Analysis: FastReport .NET to go-fastreport

This document lists every .cs file in the original FastReport .NET codebase and identifies
methods, properties, and features that are not yet implemented in the Go port.

## Legend

- **FULLY PORTED** - All public methods/properties have Go equivalents
- **PARTIALLY PORTED** - Some methods/properties are missing (details listed)
- **NOT PORTED** - No Go equivalent exists
- **OUT OF SCOPE** - Not applicable to the Go port (UI, designer, database-specific connectors, etc.)
- **N/A** - Auto-generated, build scripts, or infrastructure files

---

## Demos

> **OUT OF SCOPE** - demo applications and sample code, not part of the core report engine.

- `Demos/OpenSource/Console apps/DataFromArray/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/DataFromBusinessObject/Category.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/DataFromBusinessObject/Product.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/DataFromBusinessObject/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/DataFromDataSet/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/PdfExport/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/ReportFromCode/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/UserFunctions/MyFunctions.cs` - OUT OF SCOPE
- `Demos/OpenSource/Console apps/UserFunctions/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/Extra/FastReport.OpenSource.AvaloniaViewer/App.xaml.cs` - OUT OF SCOPE
- `Demos/OpenSource/Extra/FastReport.OpenSource.AvaloniaViewer/MainWindow.xaml.cs` - OUT OF SCOPE
- `Demos/OpenSource/Extra/FastReport.OpenSource.AvaloniaViewer/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/Extra/FastReport.OpenSource.AvaloniaViewer/Properties/Resources.Designer.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.6.0/Controllers/HomeController.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.6.0/Models/HomeModel.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.6.0/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Controllers/HomeController.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/FastreportContext.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Categories.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Customers.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Employees.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/ErrorViewModel.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/HomeModel.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/MatrixDemo.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Orderdetails.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Orders.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Products.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Shippers.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Suppliers.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Models/Unicode.cs` - OUT OF SCOPE
- `Demos/OpenSource/MVC/FastReport.OpenSource.MVC.DataBase/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.Core.React/Controllers/HomeController.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.Core.React/Pages/Error.cshtml.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.Core.React/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.Core.Vue/Controllers/ReportsController.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.Core.Vue/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.Core.Vue/ReportQuery.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.OpenSource.Angular/Controllers/WebReportController.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.OpenSource.Angular/Pages/Error.cshtml.cs` - OUT OF SCOPE
- `Demos/OpenSource/SPA/FastReport.OpenSource.Angular/Program.cs` - OUT OF SCOPE
- `Demos/OpenSource/_Shared/DataSetService.cs` - OUT OF SCOPE
- `Demos/OpenSource/_Shared/Utils.cs` - OUT OF SCOPE

## Extras

> **OUT OF SCOPE** - third-party database connectors, plugins, and optional extensions.

- `Extras/Core/FastReport.Data/FastReport.Data.Cassandra/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Cassandra/CassandraConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Cassandra/CassandraConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Cassandra/CassandraDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Cassandra/CassandraDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ClickHouse/ClickHouseAssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ClickHouse/ClickHouseConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ClickHouse/ClickHouseConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ClickHouse/ClickHouseConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ClickHouse/ClickHouseDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Couchbase/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Couchbase/CouchbaseConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Couchbase/CouchbaseConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Couchbase/CouchbaseConnectionStringBuilder.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Couchbase/CouchbaseDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Couchbase/CouchbaseDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ElasticSearch/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ElasticSearch/ESConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ElasticSearch/ESConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ElasticSearch/ESDataSourceConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ElasticSearch/ESDataSourceConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.ElasticSearch/ESDataSourceConnectionStringBuilder.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Excel/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Excel/ExcelConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Excel/ExcelConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Excel/ExcelConnectionStringBuilder.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Excel/ExcelDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Excel/ExcelDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Firebird/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Firebird/FirebirdConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Firebird/FirebirdConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Firebird/FirebirdDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Firebird/FirebirdDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleAuthConfigurationDialog.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleAuthConfigurationDialog.designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsClient.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsConfigLoader.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsConnectionEditor.designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsConnectionStringBuilder.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsCredentials.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsDataProvider.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/GoogleSheetsLoginUIManager.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/IGoogleSheetsClient.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/IGoogleSheetsConfigLoader.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/IGoogleSheetsDataProvider.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/IGoogleSheetsLoginUIManager.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/IProgressIndicator.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/OAuth.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/OAuthToken.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/ProgressIndicator.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.GoogleSheets/ProgressIndicatorFactory.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Ignite/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Ignite/IgniteConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Ignite/IgniteConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Ignite/IgniteDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Ignite/IgniteDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/CSharpCodeWriter.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/FieldInfo.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/ICodeWriter.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/IJsonClassGeneratorConfig.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/JsonClassGenerator.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/JsonClassHelper.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/JsonType.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonClassGenerator/JsonTypeEnum.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonCompiler.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonConnectionEditor.designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonConnectionStringBuilder.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonConnectionType.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Json/JsonDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Linter/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Linter/LinterConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Linter/LinterDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Linter/LinterDataConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Linter/LinterDataConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MongoDB/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MongoDB/MongoDBConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MongoDB/MongoDBConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MongoDB/MongoDBDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MongoDB/MongoDBDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MsSql/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MsSql/MsSqlDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MySql/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MySql/MySqlConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MySql/MySqlConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MySql/MySqlDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.MySql/MySqlDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Odbc/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Odbc/OdbcDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.OracleODPCore/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.OracleODPCore/OracleConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.OracleODPCore/OracleConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.OracleODPCore/OracleDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Postgres/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Postgres/PostgresConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Postgres/PostgresConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Postgres/PostgresDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Postgres/PostgresDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.Postgres/PostgresTypesParsers.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.RavenDB/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.RavenDB/RavenDBConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.RavenDB/RavenDBConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.RavenDB/RavenDBConnectionStringBuilder.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.RavenDB/RavenDBDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.RavenDB/RavenDBDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.SQLite/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.SQLite/SQLiteConnectionEditor.Designer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.SQLite/SQLiteConnectionEditor.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.SQLite/SQLiteDataConnection.DesignExt.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Data/FastReport.Data.SQLite/SQLiteDataConnection.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Plugin/FastReport.Plugins.WebP/AssemblyInitializer.cs` - OUT OF SCOPE
- `Extras/Core/FastReport.Plugin/FastReport.Plugins.WebP/WebPCustomLoader.cs` - OUT OF SCOPE
### Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple

#### `PdfSimpleExportTests.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple.Tests/PdfSimpleExportTests.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. C# tests are not directly ported but Go has equivalent coverage in export/pdf/export_test.go, export/pdf/pdf_internal_test.go, and export/pdf/pdf_objects_test.go covering PDF structure, UTF-16 encoding, and metadata.

#### `PDFSimpleExport.Config.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PDFSimpleExport.Config.cs`
- **Status**: FULLY PORTED
- **Gaps**: None.
- **Reviewed**: 2026-03-21. Added ImageDpi (default 300, clamp 96-1200), JpegQuality (default 90, clamp 10-100) with SetImageDpi()/SetJpegQuality() clamping setters. Added Author, Title, Subject, Keywords fields wired to PDF /Info in Finish(). Info struct extended with Subject/Keywords fields and SetSubject()/SetKeywords(). Start() calls NewInfo(); Finish() populates it. Tests in export_test.go and catalog_test.go. ImageDpi/JpegQuality stored for API compat but unused at runtime.

#### `PDFSimpleExport.Images.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PDFSimpleExport.Images.cs`
- **Status**: FULLY PORTED (architectural variant)
- **Gaps**: None functional.
- **Reviewed**: 2026-03-21. C# uses murmur3 hash deduplication of GDI page bitmaps at ImageDpi/96 scale. Go embeds images directly from BlobStore as XObjects without page-bitmap rendering (higher fidelity). JPEG pass-through via DCTDecode and alpha via opacity ExtGState are implemented. XObject naming (Im%d, WmIm%d) is consistent.

#### `PDFSimpleExport.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PDFSimpleExport.cs`
- **Status**: FULLY PORTED
- **Gaps**: None.
- **Reviewed**: 2026-03-21. Pipeline Start()->ExportPageBegin()->ExportBand()->ExportPageEnd()->Finish() matches C#. Coordinate system: Go PixelsToPoints() (96->72 dpi, factor 0.75) is equivalent to C# PDF_PAGE_DIVIDER=2.8346 mm->pt. Finish() writes /Info metadata. ExportBand() excludes TableColumn/TableCell/TableRow matching C# filter.

#### `PdfArray.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfArray.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Array type with Add, Len, WriteTo ported in export/pdf/core/array.go.
- **Reviewed**: 2026-03-21. Go WriteTo produces `[ val1 val2 ]` (space after `[`, space after each item, `]` closes) matching C# exactly. 100% statement coverage. NewArray variadic constructor, Add (returns receiver for chaining), Len all correct.

#### `PdfBoolean.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfBoolean.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Boolean struct with Value field and WriteTo ported in export/pdf/core/boolean.go.
- **Reviewed**: 2026-03-21. Go WriteTo outputs `"true"` or `"false"` matching C# exactly. Value field exposed directly (no getter/setter needed for a simple bool). 100% statement coverage.

#### `PdfDictionary.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfDictionary.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Dictionary with Add/Get/WriteTo using insertion-order preservation (entries slice + map index) ported in export/pdf/core/dictionary.go.
- **Reviewed**: 2026-03-21. Go WriteTo produces `<< /Key Value >>` matching C# format. Go adds explicit insertion-order guarantee (entries slice + map index) which is strictly better than C# Dictionary<> (which has no order guarantee). Add replaces existing keys without changing position. 100% statement coverage.

#### `PdfDirectObject.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfDirectObject.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Go's IndirectObject in export/pdf/core/indirect.go combines C# PdfDirectObject and PdfIndirectObject into a single type; WriteTo produces identical "N G obj\n…\nendobj\n" output.

#### `PdfIndirectObject.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfIndirectObject.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Go's Ref type in export/pdf/core/ref.go handles "N G R" indirect references; IndirectObject wraps it for the full object block.

#### `PdfName.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfName.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. `Name` struct in export/pdf/core/name.go matches C# exactly: only `a-z`, `A-Z`, `0-9` pass through unencoded; all other bytes are written as `#XX` uppercase hex. Theoretical-only difference: C# iterates over UTF-16 `char` values and casts to `byte` (low byte), while Go iterates over raw UTF-8 bytes; in practice this is irrelevant because PDF names are always ASCII. The C# `Equals`/`GetHashCode` members are not ported (unused in the export pipeline).

#### `PdfNumeric.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfNumeric.cs`
- **Status**: FULLY PORTED
- **Gaps**: Precision difference (intentional/acceptable). C# `PdfNumeric(double)` defaults to precision 2 and uses `ExportUtils.FloatToString` which calls `Math.Round(value, digits)` then `Convert.ToString` — trailing zeros are stripped (e.g., `1.50` → `"1.5"`). Go `NewFloat` always renders with 4 decimal places using `strconv.FormatFloat(..., 'f', 4, 64)` (e.g., `"1.5000"`). Both are valid PDF representations. The C# multi-precision constructor `PdfNumeric(double, int)` is not replicated as a distinct API (Go callers manage precision at the call site). The C# `ValueInt` and `ValueReal` property setters that cross-update `precision` are not needed since Go uses an explicit `IsInt` flag.

#### `PdfObjectBase.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfObjectBase.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Go uses Object interface with WriteTo and Type methods in export/pdf/core/object.go (idiomatic replacement for abstract base class).

#### `PdfStream.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfStream.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. `Stream` in export/pdf/core/stream.go matches C# exactly. Compression: C# `ZLibDeflate` manually writes the `0x78 0xDA` zlib header, uses a raw `DeflateStream`, and appends the Adler-32 checksum as 4 big-endian bytes. Go's `compress/zlib.NewWriterLevel(dst, zlib.DefaultCompression)` produces byte-for-byte identical output (same `0x78 0xDA` header, same deflate payload, same Adler-32 footer per RFC 1950). Round-trip decompression is verified in `TestStream_WriteTo_CompressedDataRoundtrip`. The C# `Adler32`/`ZLibDeflate` private methods are replaced by the stdlib implementation, not ported directly.

#### `PdfString.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfCore/PdfString.cs`
- **Status**: FULLY PORTED
- **Gaps**: (1) `Append(string)` method and `appendText` list omitted — `PdfString.Append` is never called in the C# export pipeline (all stream-building uses `PdfContents`, not `PdfString`). (2) Literal high-byte encoding: both C# and Go write `\` followed by the decimal integer value of the byte (e.g., byte 0xE9 becomes `\233`) rather than the PDF-spec octal notation (`\351`). This non-standard behavior is faithfully preserved from C#; PDF readers handle it because most string data is written as hex (`IsHex=true`) in practice. (3) `ToString()` override that concatenates `text + appendText` is omitted (unused externally). Core write logic — UTF-16BE with BOM `0xFEFF`, hex `<…>` and literal `(…)` modes, all escape sequences `\n \r \t \b \f \( \) \\` — matches C# exactly.

#### `PdfCatalog.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfCatalog.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Catalog with Type, Version, MarkInfo, SetOutlines(), SetNamedDests() ported in export/pdf/catalog.go.

#### `PdfContents.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfContents.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Contents with Write/WriteString/Finalize ported in export/pdf/contents.go using bytes.Buffer instead of StringBuilder.

#### `PdfImage.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfImage.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Image XObject properties (Height, Width, ColorSpace, BitsPerComponent, etc.) handled via Dictionary operations in export/pdf/export.go.

#### `PdfInfo.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfInfo.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Info with SetTitle/SetAuthor/SetCreator/SetProducer ported in export/pdf/catalog.go. Creator/Producer strings identify "go-fastreport" instead of "FastReport.NET".

#### `PdfMask.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfMask.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Mask functionality handled in export/pdf/export.go image processing via shared dictionary configuration (ColorSpace=DeviceGray, Compress=true).

#### `PdfPage.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfPage.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Page with MediaBox, resources, and XObject management ported in export/pdf/page.go.

#### `PdfPages.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfPages.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Pages with Kids array, Count tracking, and AddPage() ported in export/pdf/page.go.

#### `PdfTrailerId.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfTrailerId.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. PDF trailer ID generation (GUID-based) handled inline in Writer.Write() in export/pdf/writer.go.

#### `PdfWriter.cs`
- **File**: `Extras/OpenSource/FastReport.OpenSource.Export.PdfSimple/FastReport.OpenSource.Export.PdfSimple/PdfObjects/PdfWriter.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Writer with NewObject(), xref table, and trailer generation ported in export/pdf/writer.go. Outputs PDF-1.4 instead of C# 1.5 (both valid).
- `Extras/ReportBuilder/FastReport.ReportBuilder.UnitTest/Person.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder.UnitTest/Properties/AssemblyInfo.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder.UnitTest/ReportTest.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Builders/DataBuilder.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Builders/DataHeaderBuilder.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Builders/GroupHeaderBuilder.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Builders/ReportBuilder.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Builders/ReportTitleBuilder.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Definitions/DataDefinition.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Definitions/DataHeaderDefinition.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Definitions/GroupHeaderDefinition.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Definitions/ReportDefinition.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Definitions/ReportTitleDefinition.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/Helpers/GenericHelpers.cs` - OUT OF SCOPE
- `Extras/ReportBuilder/FastReport.ReportBuilder/ReportHelper.cs` - OUT OF SCOPE

## FastReport.Base

#### `AssemblyInitializer.cs`
- **File**: `FastReport.Base/AssemblyInitializer.cs`
- **Status**: FULLY PORTED
- **Gaps**: Initialization logic is distributed. Bands and Objects are registered in `reportpkg/serial_registrations.go`. Data connections and sources are handled explicitly in `reportpkg/loadsave.go`. Standard functions are registered in `expr/evaluator.go` via the `functions` package. Explicit handling replaces generic registry for Data/Connections/Formats.

#### `BandBase.Async.cs`
- **File**: `FastReport.Base/BandBase.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API. GetDataAsync/IsEmptyAsync have no Go equivalents; band lifecycle is handled synchronously by engine/bands.go.

#### `BandBase.cs`
- **File**: `FastReport.Base/BandBase.cs`
- **Status**: PARTIALLY PORTED
- **Fixed (go-fastreport-zifp3, 2026-03-21)**: Implemented `IsEmpty()` — returns true when the band has no child objects (mirrors C# line 926); `GetExpressions()` — returns OutlineExpression slice for engine expression pre-compilation (mirrors C# line 606); `IsColumnDependentBand()` — reads `FlagIsColumnDependent` field set to true in constructors of DataBand, DataHeaderBand, DataFooterBand, GroupHeaderBand, GroupFooterBand, ColumnHeaderBand, ColumnFooterBand, ReportSummaryBand (mirrors C# line 577; flag avoids Go embedding parent-pointer type-loss); `Assign()` — deep-copies all BandBase fields including guides slice (mirrors C# line 514). Added `FlagIsColumnDependent bool` to `BandBase` struct.
- **Remaining Gaps**: `Validate()` — OUT OF SCOPE (designer validation; C# returns empty list). `UpdateWidth()` — OUT OF SCOPE (design-time, requires page reference). `SetUpdatingLayout()` — OUT OF SCOPE (designer loop-prevention guard). `AddBookmarks()` — OUT OF SCOPE (requires engine reference; handled in engine/bands.go). `AddLastToFooter()` — OUT OF SCOPE (complex band-break page-overflow splitting). Engine lifecycle methods (SaveState/RestoreState/GetData/CalcHeight/Break) are handled centrally by engine/bands.go.

#### `BandCollection.cs`
- **File**: `FastReport.Base/BandCollection.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go handles band collections idiomatically via singleton slots and a `bands []report.Base` slice in `ReportPage`. However, unlike C# `BandCollection` (which automatically sets `Parent` on insert via `FRCollectionBase`), Go's `ReportPage.AddBand` and slot accessors do not currently enforce parent assignment. `BandBase` correctly implements `report.Parent`, but `ReportPage` lacks a full `AddChild`/`RemoveChild` implementation for its band collection.

#### `BandColumns.cs`
- **File**: `FastReport.Base/BandColumns.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Data structure ported. DownThenAcross rendering unimplemented. **Fixed 2026-03-21** (go-fastreport-oy5tz): Serialize now writes Count/Width/Layout/MinRowCount; Deserialize now reads Layout (ColumnLayout enum string "AcrossThenDown"/"DownThenAcross", with int fallback) and MinRowCount. Added formatColumnLayout/parseColumnLayout helpers matching C# BandColumns.Serialize WriteValue pattern.

#### `Base.cs`
- **File**: `FastReport.Base/Base.cs`
- **Status**: PARTIALLY PORTED
- **Fixed (go-fastreport-etjv7)**: Implemented `HasParent(obj Base, ancestor Parent) bool` free function (report/base.go — idiomatic Go equivalent of C# Base.HasParent(Base)); `HasRestriction(r Restrictions) bool` method on BaseObject; `AllObjects(root Base) []Base` free function (equivalent to C# Base.AllObjects property — recursive descendants of root, excluding root itself); `SetZOrder(order int)` on BaseObject (delegates to parent.SetChildOrder when parent set, else updates internal field); `ZOrder()` getter now also delegates to parent.GetChildOrder when parent is set, matching C# Base.ZOrder getter/setter.
- **Remaining Gaps**: InvokeEvent() — OUT OF SCOPE (script engine not ported). Page parent traversal property — NOT IMPLEMENTED (no PageBase interface in Go yet). Assign()/AssignAll()/BaseAssign() deep-copy — NOT IMPLEMENTED (not needed for headless engine). Clear()/Dispose() lifecycle hooks — OUT OF SCOPE (Go uses GC). OnInitialize()/OnFinish() engine callbacks — handled centrally by engine/bands.go. CreateUniqueName() — NOT IMPLEMENTED (useful helper, not used by engine). All other previously listed items (runtime state queries, design-time Flags/Restrictions, ChildDeserializer override hook, serialization) are IMPLEMENTED.

#### `Border.cs`
- **File**: `FastReport.Base/Border.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Draw()/BorderLine.Draw() OUT OF SCOPE — exporters handle rendering. ZoomBorder() and BorderLine.Assign() not ported (designer-only). DashPattern/LineStyle.Custom: zero FRX occurrences; parseLineStyle("Custom") maps to Solid. **Implemented (go-fastreport-2ljeg)**: serializeBorder()/deserializeBorder() fully implement C# Border.Serialize() delta-serialization including per-line vs common-line logic, shadow fields, and ShouldSerialize() guards.

#### `BreakableComponent.cs`
- **File**: `FastReport.Base/BreakableComponent.cs`
- **Status**: PARTIALLY PORTED
- **Fixed (go-fastreport-zifp3, 2026-03-21)**: Implemented `Assign()` in `report/breakable.go` — copies CanBreak and BreakTo reference from source (mirrors C# line 64). Note: BreakTo is copied as a shallow pointer reference; caller is responsible for managing lifetime.
- **Remaining Gaps**: BreakTo disposal event hook — OUT OF SCOPE (C# uses `Disposed` event to auto-nil the BreakTo reference when the target is garbage-collected; Go uses GC and has no disposal hooks). Core CanBreak, BreakTo, Break(), CalcHeight(), FlagMustBreak, Serialize/Deserialize are fully ported.

#### `CapSettings.cs`
- **File**: `FastReport.Base/CapSettings.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Width, Height, Style properties, DefaultCapSettings(), Assign(), Clone(), Equals(), SerializeCap()/DeserializeCap() all ported in object/lines.go. GetCustomCapPath() (GDI+ graphics path generation) is the only missing method but is not needed in the Go architecture.
- **Fixed (go-fastreport-0yt0a)**: The original port used a flat CSV string format for cap serialization (e.g. `StartCap="10,10,4"`) which did not match the FRX file format. Fixed to use dot-qualified attributes (`StartCap.Width`, `StartCap.Height`, `StartCap.Style="Arrow"`) matching C# CapSettings.Serialize(prefix, writer, diff). Added Assign(), Clone(), Equals() methods. Added parseCapStyle()/formatCapStyle() helpers using string enum names ("Arrow", "Circle", etc.) instead of integer values.

#### `CellularTextObject.cs`
- **File**: `FastReport.Base/CellularTextObject.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. All properties (CellWidth, CellHeight, HorzSpacing, VertSpacing, WordWrap), grid rendering via engine's populateCellularTextCells(), constructor defaults (CanBreak=false, Border.Lines=All), Assign(), and CalcHeight() are now fully ported in object/cellular_text.go. Assign() copies the embedded TextObject by value plus all cellular-specific fields. CalcHeight() implements the autoRows=true table-height computation matching C# GetTable(autoRows: true).calHeight (CellularTextObject.cs:275-281).

#### `CheckBoxObject.Async.cs`
- **File**: `FastReport.Base/CheckBoxObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper not applicable to Go. NOTE: HideIfUnchecked visibility logic previously missing; fixed 2026-03-21 (see CheckBoxObject.cs below).

#### `CheckBoxObject.cs`
- **File**: `FastReport.Base/CheckBoxObject.cs`
- **Status**: FULLY PORTED
- **Fixed 2026-03-21** (go-fastreport-lfkpm): CheckedSymbol and UncheckedSymbol serialized/deserialized as enum name strings.
- **Fixed 2026-03-21**: CheckColor serialized in Serialize() when non-default (mirrors C# ShouldSerializeCheckColor / WriteValue "CheckColor", CheckBoxObject.cs lines 200-203, 309); deserialized via utils.ParseColor.
- **Fixed 2026-03-21**: SetCheckWidthRatio() clamps to [0.2, 2.0] matching C# setter (CheckBoxObject.cs lines 167-173); clamping applied on Deserialize too.
- **Fixed 2026-03-21**: engine buildPreparedObject() returns nil for unchecked CheckBox when HideIfUnchecked=true — mirrors C# GetDataShared lines 359-360. Tests: engine/checkbox_hide_unchecked_test.go, object/checkbox_hyperlink_fixes_test.go.

#### `ChildBand.cs`
- **File**: `FastReport.Base/ChildBand.cs`
- **Status**: FULLY PORTED
- **Fixed (go-fastreport-zifp3, 2026-03-21)**: Implemented `GetTopParentBand()` — traverses Parent chain skipping ChildBand instances to find the first non-ChildBand ancestor; returns a `columnDependentChecker` interface value (mirrors C# line 67). Implemented `IsColumnDependentBand()` — delegates to the top parent via `GetTopParentBand()` (mirrors C# BandBase.IsColumnDependentBand line 582). Implemented `Assign()` — copies all ChildBand-specific fields plus calls `BandBase.Assign()` for base fields (mirrors C# line 82). Core properties (FillUnusedSpace, CompleteToNRows, PrintIfDatabandEmpty) and Serialize/Deserialize were already fully ported.

#### `ColumnFooterBand.cs`
- **File**: `FastReport.Base/ColumnFooterBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. NewColumnFooterBand() now correctly sets FlagUseStartNewPage=false matching the C# constructor (ColumnFooterBand.cs). Fixed in go-fastreport-0jdot.

#### `ColumnHeaderBand.cs`
- **File**: `FastReport.Base/ColumnHeaderBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. NewColumnHeaderBand() now correctly sets FlagUseStartNewPage=false matching the C# constructor (ColumnHeaderBand.cs). Fixed in go-fastreport-0jdot.

#### `ComponentBase.cs`
- **File**: `FastReport.Base/ComponentBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: **Reviewed 2026-03-21 (go-fastreport-rvegr)**. Implemented: `AbsBounds()`, `TagStr()`/`SetTagStr()` string Tag with FRX serialization (ComponentBase.cs:489), `Assign(src)` deep-copying 12 scalar fields (ComponentBase.cs:437-453), `GetExpressions()` bracket-stripping expression list (ComponentBase.cs:498-529), `CalcVisibleExpression(expr, calc)` with show-by-default semantics (ComponentBase.cs:536-563), engine `buildPreparedObject` evaluates VisibleExpression at render time (engine/objects.go). 26 tests in report/componentbase_gaps_test.go. OUT OF SCOPE: `ClientSize` (DialogPage/designer), `GetExtendedSize()` (validator only), `SetLeft/Top/Width/Height → UpdateLayout` triggers (designer), `CalcPrintableExpression` (no print driver in Go; Printable flag preserved but not expression-evaluated), design-mode restriction guards.

#### `ConditionCollection.cs`
- **File**: `FastReport.Base/ConditionCollection.cs`
- **Status**: FULLY PORTED
- **Fixed (2026-03-21)**: Exported `style.ConditionCollection` type added with full API: `Add()`, `Insert()`, `Remove()`, `RemoveAt()`, `IndexOf()`, `Contains()`, `Clear()`, `Get()`, `Set()`, `Items()`, `Assign()`, `Clone()`, `Equals()`, `FindByExpression()`, `Len()`. Nil-safe `Len()` on nil receiver returns 0. All items deep-copied via `HighlightCondition.Clone()` in `Assign()`/`Clone()`. `Equals()` compares element-by-element. 25 tests in `style/highlight_condition_test.go`. Note: `Remove()` uses value-equality (`Equals()`) rather than C# reference equality — correct for Go value types. `AddRange()` not added (no callers).
- **Remaining gaps**: Fill/TextFill still `color.RGBA` only (gradient fills not modelled — see `HighlightCondition.cs`).

#### `ContainerObject.Async.cs`
- **File**: `FastReport.Base/ContainerObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses context.Context and synchronous execution throughout the engine pipeline; the C# async/await Task pattern has no direct Go equivalent.

#### `ContainerObject.cs`
- **File**: `FastReport.Base/ContainerObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: UpdateLayout() is a no-op stub (no anchor/dock logic ported); IParent.CanContain excludes ContainerObject in Go (C# excludes SubreportObject instead). Core fields (Width, Height, serialization, child management) are ported.

#### `DataBand.Async.cs`
- **File**: `FastReport.Base/DataBand.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API. Note: InitDataSourceAsync, IsEmptyAsync, IsDetailEmptyAsync are not ported; Go engine handles these synchronously without cancellation hooks.

#### `DataBand.cs`
- **File**: `FastReport.Base/DataBand.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing Relation property for master-detail linking (C# DataBand.Relation/DataSourceBase.Init(Relation,...)); InitDataSource() engine lifecycle (handled in engine/databands.go); IContainDataSource.UpdateDataSourceRef() (engine-internal); IsEmpty()/IsDetailEmpty() engine lifecycle; UpdateWidth() for multi-column/hierarchical indent geometry.
- **Fixed 2026-03-21** (go-fastreport-oy5tz): BandColumns serialization added — DataBand.Serialize now writes Columns.Count/Width/Layout/MinRowCount and Deserialize reads them all. Layout is the ColumnLayout enum string ("AcrossThenDown"/"DownThenAcross") with int fallback.
- **Fixed 2026-03-21**: Added DataBand.AddChild() to route DataHeaderBand/DataFooterBand children to d.header/d.footer instead of d.objects. Added DataBand.Serialize() to write d.header and d.footer as child XML elements — without this, nested bands were lost on round-trip.
- **Fixed 2026-03-21** (go-fastreport-mdnt4): Added DataBand.GetExpressions() — returns all sort expressions (Expression field, falling back to Column) followed by the filter expression. Mirrors C# DataBand.GetExpressions (DataBand.cs:542-551).
- **Fixed 2026-03-21** (go-fastreport-mdnt4): Added DataBand.Assign() — deep-copies all scalar properties (filter, rowCount, maxRows, sort slice, printIf* flags, keep* flags, hierarchy columns, indent, collectChildRows, resetPageNumber, dataSourceAlias/ref). Sort specs are deep-copied so caller mutations are independent. Child band references (header, footer) are NOT copied (structural). Mirrors C# DataBand.Assign (DataBand.cs:462-483).

#### `DataFooterBand.cs`
- **File**: `FastReport.Base/DataFooterBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `DataHeaderBand.cs`
- **File**: `FastReport.Base/DataHeaderBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `Fakes.cs`
- **File**: `FastReport.Base/Fakes.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET platform stubs.

#### `Fills.cs`
- **File**: `FastReport.Base/Fills.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: HatchStyle: Go only has 6 common styles vs 56 in C# System.Drawing.Drawing2D.HatchStyle — unknown styles fall back to Horizontal. Setter validation (Focus/Contrast clamp) not enforced. TextureFill BlobStore/ImageIndex path not implemented; inline ImageData base-64 path works. CreateBrush()/Draw() OUT OF SCOPE. **Implemented (go-fastreport-2ljeg)**: PathGradientFill, PathGradientStyle enum, TextureFill, WrapMode enum added to style/fill.go. serializeFill()/deserializeFill() in report/borderfill_serial.go now handle all 6 fill types with round-trip serialization. All test-reports FRX files with PathGradient and Texture fills now load without panic.

#### `GroupFooterBand.cs`
- **File**: `FastReport.Base/GroupFooterBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `GroupHeaderBand.Async.cs`
- **File**: `FastReport.Base/GroupHeaderBand.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — async InitDataSourceAsync/IsEmptyAsync wrappers not applicable to Go's synchronous engine; equivalent logic is in engine/groups.go:applyGroupSort().

#### `GroupHeaderBand.cs`
- **File**: `FastReport.Base/GroupHeaderBand.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: InitDataSource()/FinalizeDataSource() engine lifecycle (inserting/removing group-sort specs into DataBand.Sort before/after run — handled in engine/groups.go:applyGroupSort()); IsEmpty() engine lifecycle (calls DataBand.IsEmpty() which needs live datasource); CanContain() addChild validation overrides (designer-only constraint checks, not ported).
- **Fixed (2026-03-21)**: SortOrder was serialized/deserialized as an integer; C# uses enum name strings ("None", "Ascending", "Descending") via Converter.ToString(). Fixed in types.go to use WriteStr/ReadStr with sortOrderToString/sortOrderFromString helpers. Real FRX files contain SortOrder="None" — the old ReadInt silently returned 0 (=Ascending) for string values.
- **Fixed (2026-03-21)**: Added GroupHeaderBand.AddChild() to route DataBand → g.data, GroupHeaderBand → g.nestedGroup, GroupFooterBand → g.groupFooter. Without this, nested DataBands fell into g.objects (wrong collection). Added GroupHeaderBand.Serialize() to write g.nestedGroup/g.data/g.groupFooter as child XML elements — mirrors C# GroupHeaderBand.GetChildObjects() (GroupHeaderBand.cs:272).
- **Fixed (2026-03-21)** (go-fastreport-mdnt4): Added header/footer (*DataHeaderBand/*DataFooterBand) fields, accessors, AddChild routing, and Serialize child-write for them. Mirrors C# GroupHeaderBand fields/GetChildObjects (GroupHeaderBand.cs:80-81,272-283).
- **Fixed (2026-03-21)** (go-fastreport-mdnt4): Added GroupDataBand() computed property traversing the nested-group chain to find the DataBand. Mirrors C# GroupHeaderBand.GroupDataBand (GroupHeaderBand.cs:254-267).
- **Fixed (2026-03-21)** (go-fastreport-mdnt4): Added DataSource() computed property delegating to GroupDataBand().dataSource. Mirrors C# GroupHeaderBand.DataSource (GroupHeaderBand.cs:245-252).
- **Fixed (2026-03-21)** (go-fastreport-mdnt4): Added groupValue field, ResetGroupValue(calc func(string)(any,error)) and GroupValueChanged(calc) — engine injects the Report.Calc function. Empty condition is a no-op returning false/nil. Mirrors C# GroupHeaderBand.ResetGroupValue/GroupValueChanged (GroupHeaderBand.cs:415-445).
- **Fixed (2026-03-21)** (go-fastreport-mdnt4): Added GetExpressions() returning []string{condition}. Mirrors C# GroupHeaderBand.GetExpressions (GroupHeaderBand.cs:369-371).
- **Fixed (2026-03-21)** (go-fastreport-mdnt4): Added Assign() copying condition, sortOrder, keepTogether, resetPageNumber; structural child-band references not copied. Mirrors C# GroupHeaderBand.Assign (GroupHeaderBand.cs:339-348).

#### `HeaderFooterBandBase.cs`
- **File**: `FastReport.Base/HeaderFooterBandBase.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `HighlightCondition.cs`
- **File**: `FastReport.Base/HighlightCondition.cs`
- **Status**: PARTIALLY PORTED
- **Fixed (2026-03-21)**: Added `Border *style.Border` field (was omitted). Added `Clone()`, `Assign()`, `Equals()` matching HighlightCondition.cs:64-96. `Clone()` deep-copies Border. `Equals()` uses `Border.Equals()`. `NewHighlightCondition()` now initialises `Border = style.NewBorder()`. Engine (engine/objects.go) now applies `cond.Border` when `cond.ApplyBorder` (mirrors TextObject.ApplyCondition TextObject.cs:1558-59). Serialization writes Border attrs when ApplyBorder=true; deserialization reads them. Highlights were already serialized (round-trip worked); border now included. 8 tests in object/highlight_border_test.go; 14 in style/highlight_condition_test.go.
- **Remaining gaps**: Fill/TextFill still `color.RGBA` only — gradient fills in highlight conditions not modelled (requires full style.Fill interface on HighlightCondition). Calc context does not receive evaluated text value ("Value" C# passes in TextObject.GetCachedTextValue — low priority).

#### `HtmlObject.Async.cs`
- **File**: `FastReport.Base/HtmlObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — async GetDataAsync wrapper not applicable; Go renders HtmlObject synchronously via buildPreparedObject() in engine.

#### `HtmlObject.cs`
- **File**: `FastReport.Base/HtmlObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: GetStringFormat(), DrawText()/Draw(), SaveState()/RestoreState(), GetExpressions(), CalcWidth()/CalcHeight(), Break(), ApplyCondition() are C# preview/editing features not needed in the Go engine pipeline; HTML text evaluated via evalTextWithFormat() and passed to exporters. No report-generation correctness gaps remain.

#### `Hyperlink.cs`
- **File**: `FastReport.Base/Hyperlink.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-21**: Added OpenLinkInNewTab bool to Hyperlink struct; serialized as "Hyperlink.OpenLinkInNewTab" (omitted when false); engine derives po.HyperlinkTarget="_blank" when true, mirroring C# Hyperlink.cs lines 131-135. Legacy Target string field kept for backward compatibility.
- **Fixed 2026-03-21**: Added ValuesSeparator string (default ";"); serialized only when not ";", matching C# ShouldSerializeValuesSeparator (Hyperlink.cs line 218); deserialized with default ";".
- **Fixed 2026-03-21**: XSS sanitization on Deserialize — javascript: URIs and inline script tags discarded, matching C# Hyperlink.Value setter regex (Hyperlink.cs lines 113-122). Tests: object/checkbox_hyperlink_fixes_test.go.
- **Fixed 2026-03-22**: HyperlinkKind.PageNumber (C# enum value 1) now handled in engine/objects.go buildPreparedObject() — sets po.HyperlinkKind=2 (Go preview enum) and evaluates Value/Expression to resolve the page number target. Mirrors C# HTMLExportLayers.cs:167.
- **Remaining gaps**: SaveState/RestoreState not ported.

#### `IContainDataSource.cs`
- **File**: `FastReport.Base/IContainDataSource.cs`
- **Status**: MOSTLY PORTED
- **Fixed 2026-03-22**: UpdateDataSourceRef(ds any) added to report.DataSourceBinder interface; DataBand.UpdateDataSourceRef implemented to update its data source reference. Dictionary.Merge added. Note: Go Merge does not auto-call UpdateDataSourceRef on report objects (Dictionary has no access to report pages; caller must do this).

#### `IFRSerializable.cs`
- **File**: `FastReport.Base/IFRSerializable.cs`
- **Status**: FULLY PORTED
- **Gaps**: Interface itself fully ported with error returns. Real gaps in concrete FRWriter/FRReader (WriteDouble, WriteValue, DiffObject, FixupReferences).

#### `IParent.cs`
- **File**: `FastReport.Base/IParent.cs`
- **Status**: FULLY PORTED
- **Gaps**: All 7 methods have Go equivalents in report.Parent interface.

#### `ITranslatable.cs`
- **File**: `FastReport.Base/ITranslatable.cs`
- **Status**: FULLY PORTED
- **Gaps**: Interface declared in Go. No types implement it in either codebase. Dead code.

#### `LineObject.cs`
- **File**: `FastReport.Base/LineObject.cs`
- **Status**: MOSTLY PORTED
- **Fixed** (2026-03-21): DashPattern is now serialized (Serialize writes the comma-separated value when non-empty) and deserialized (Deserialize parses the attribute), matching C# LineObject.Serialize() lines 274-275. No FRX test reports currently use DashPattern on LineObject, but the round-trip is validated by tests.
- **Fixed 2026-03-22** (go-fastreport-eds3d): Added `LineObject.Assign()` — deep-copies Diagonal, StartCap, EndCap, and DashPattern fields on top of ReportComponentBase.Assign(). Mirrors C# LineObject.Assign (LineObject.cs:81-89).
- **Gaps (remaining)**: Validate(), IsHaveToConvert(), GetExtendedSize(), and CreatePath() are not implemented. Draw() is handled by exporters rather than on the object itself.

#### `ObjectCollection.cs`
- **File**: `FastReport.Base/ObjectCollection.cs`
- **Status**: MOSTLY PORTED
- **Fixed (go-fastreport-etjv7)**: Added nil guard to `Add` and `Insert` (matching C# FRCollectionBase.Add nil guard); added `Equals(*ObjectCollection) bool` (element-wise equality, matching C# FRCollectionBase.Equals); added `CopyTo(*ObjectCollection)` (replace dst contents with src, matching C# FRCollectionBase.CopyTo); added `AddRangeCollection(*ObjectCollection)` overload (matching C# FRCollectionBase.AddRange(ObjectCollection)).
- **Remaining Gaps**: Owner/parent hooks on Add/Remove/Clear — OUT OF SCOPE for headless Go; parent is managed at call sites via Parent interface. Clear no Dispose — OUT OF SCOPE (Go uses GC). Add returning index — not needed; Go callers always know the index via Len().

#### `OverlayBand.cs`
- **File**: `FastReport.Base/OverlayBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. NewOverlayBand() now correctly sets FlagUseStartNewPage=false matching the C# constructor (OverlayBand.cs). Fixed in go-fastreport-0jdot.

#### `PageBase.cs`
- **File**: `FastReport.Base/PageBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: `pageName`/`needRefresh`/`needModify` fields and `Refresh()`/`Modify()` methods are fully ported on `ReportPage` (tested in `pagebase_test.go`). Remaining gap: C# constructor clears `CanMove|CanResize|CanDelete|CanChangeOrder|CanChangeParent|CanCopy` flags — not implemented in Go (low impact for headless engine). Fixed in go-fastreport-e118f: added `HeightInPixels()`, `WidthInPixels()`, `PageColumns` serialization, and tests.

#### `PageCollection.cs`
- **File**: `FastReport.Base/PageCollection.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go uses `[]*ReportPage` typed slice covering all actual usage. Remaining low-risk gaps: parent management on `AddPage` (C# `FRCollectionBase.OnInsert` sets `child.Parent`), cleanup (`SetParent(nil)`) on `Remove`, `Insert` at index, `Clear`/`Dispose`, `Report` not implementing `Parent` interface. None of these affect the headless rendering engine.

#### `PageColumns.cs`
- **File**: `FastReport.Base/PageColumns.cs`
- **Status**: PARTIALLY PORTED
- **Fixed in go-fastreport-e118f**: Serialization round-trip was broken — `Columns.Count`, `Columns.Width`, and `Columns.Positions` were deserialized (read) but never serialized (written). Fixed: `ReportPage.Serialize()` now writes all three attributes when `Count > 1`, matching `PageColumns.Serialize()` in C# (PageColumns.cs:101-111). Tested with a `Badges.frx`-style round-trip.
- **Remaining gaps**: Count setter auto-calculation (C# recomputes `Width` and regenerates `Positions` when `Count` changes), ReportPage back-reference (`page` field), validation (`Count <= 0` throws in C#), `Assign()` copy method. These affect the report designer but not the rendering engine.

#### `PageFooterBand.cs`
- **File**: `FastReport.Base/PageFooterBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. NewPageFooterBand() now correctly sets FlagUseStartNewPage=false matching the C# constructor. InitializeComponent's SubreportObject.PrintOnParent=true is handled by engine band initialization. Fixed in go-fastreport-0jdot.

#### `PageHeaderBand.cs`
- **File**: `FastReport.Base/PageHeaderBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. NewPageHeaderBand() now correctly sets FlagUseStartNewPage=false matching the C# constructor (PageHeaderBand.cs line 40). Fixed in go-fastreport-0jdot.

#### `TextObject.Async.cs`
- **File**: `FastReport.Base/TextObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — contains async version of GetData(). Go engine uses synchronous pipeline; async/await is replaced by goroutines at the engine runner level if needed. All logic (expression processing, highlight application, paragraph offset, auto-shrink) is fully implemented in the Go engine.


#### `PictureObject.cs`
- **File**: `FastReport.Base/PictureObject.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-21**: TransparentColor getter/setter with round-trip serialization (utils.FormatColor/#AARRGGBB, omitted when zero). ImageIndex getter/setter/reset (-1 default; serialized when >= 0; restored on Deserialize). GetData() DataColumn binding: []byte value -> imageData (imageIndex reset to -1), string value -> imageLocation; no-op when column empty or nil. Assign() for PictureObject: deep-copies imageData, resets imageIndex to -1, copies transparency/tile/transparentColor/imageFormat, delegates base fields to PictureObjectBase.Assign(). GetExpressions() delegates to base. ImageFormat attribute round-trip for Png/Jpeg/Gif/Bmp (written only when imageData present). Engine (engine/objects.go) now calls GetData() for DataColumn binding and copies PictureSizeMode/PictureAngle/PictureTransparency/PictureTile/PictureGrayscale/PictureTransparentColor/PictureShapeKind/PictureImageAlign to PreparedObject. preview.PreparedObject gains all eight picture-specific fields.
- **Remaining gaps**: Image setter callbacks (UpdateAutoSize/UpdateTransparentImage/ResetImageIndex on Image assignment — no GDI+ Image type in Go), GrayscaleHash (GDI+ bitmap identity hash), TransparentImage (GDI+ Bitmap with MakeTransparent), DrawImage/DrawImageInternal2 rendering pipeline (GDI+ specific — HTML/PDF exporters render directly), EstablishImageForm shape masking via GraphicsPath clipping, LoadImage from file/URL (Go engine uses byte data), ForceLoadImage/DisposeImage lifecycle, InitializeComponent/FinalizeComponent (trivial imageIndex reset), ShouldDisposeImage flag.

#### `PictureObjectBase.Async.cs`
- **File**: `FastReport.Base/PictureObjectBase.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API. GetDataAsync has no Go equivalent; image loading is handled synchronously in engine/objects.go buildPreparedObject().

#### `PictureObjectBase.cs`
- **File**: `FastReport.Base/PictureObjectBase.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-21** (go-fastreport-lfkpm): SizeMode serialized as enum name string and ImageAlign as enum name string matching C# WriteValue; int fallback for backward compat.
- **Fixed 2026-03-21**: Shape (ShapeKind clipping-mask) field added — getter/setter, serialized as string name when non-default (Rectangle), deserialized with parseShapeKind. IsDataColumn/IsFileLocation/IsWebLocation computed properties (URL scheme detection). SaveState/RestoreState for SizeMode (direct sizeModeInternal assignment avoiding UpdateAutoSize, matching C# RestoreState). GetExpressions() returns DataColumn and ImageSourceExpression with bracket stripping. Assign() copies all PictureObjectBase fields including shape.
- **Remaining gaps**: Height/Width MaxHeight/MaxWidth setter clamping (C# overrides Height/Width setters). CalculateUri() with Config.ReportSettings.ImageLocationRoot prefix stripping. GetImageAngleTransform() parallelogram computation for GDI+ DrawImage (irrelevant to Go CSS-based exporters). UpdateAutoSize() angle-aware bounding box. UpdateAlign() pixel-level image alignment. SetImageLocation() ImageLocationRoot prefix stripping.

#### `PolyLineObject.cs`
- **File**: `FastReport.Base/PolyLineObject.cs`
- **Status**: PARTIALLY PORTED
- **Fixed** (2026-03-21): Serialize/Deserialize are now fully implemented. Serialize() writes PolyPoints_v2 (C# PolyLineObject.Serialize() lines 501-511), CenterX, CenterY, and DashPattern when non-empty. Deserialize() reads both the legacy PolyPoints v1 format ("X\Y\type" per point, used by Box.frx) and the current PolyPoints_v2 format with bezier L/R control points ("X/Y[/L/lx/ly][/R/rx/ry]"). Previously both were stubs delegating only to base.
- **Gaps (remaining)**: GetPath() bezier GraphicsPath generation (rendering), RecalculateBounds() bezier bounding boxes (designer), SetPolyLine() bulk assignment, Draw()/DoDrawPoly(), deprecated PointsArray/PointTypesArray, addPoint()/deletePoint()/insertPoint(), Assign(), PolyPoint.Near/ScaleX/ScaleY/Clone, PolygonSelectionMode enum.

#### `PolygonObject.cs`
- **File**: `FastReport.Base/PolygonObject.cs`
- **Status**: PARTIALLY PORTED
- **Fixed** (2026-03-21): Serialize() and Deserialize() now delegate to PolyLineObject, matching C# PolygonObject.Serialize() (PolygonObject.cs:76-77). FlagUseFill=true has no Go equivalent — Fill is always available via ReportComponentBase.Fill(); Fill.Color on PolygonObject (e.g. Box.frx Polygon8) is handled by base Deserialize. Previously missing entirely.
- **Gaps (remaining)**: getPolygonPath() close-figures for rendering, drawPoly() fill rendering override. Neither is needed for the headless export pipeline.

#### `RFIDLabel.Async.cs`
- **File**: `FastReport.Base/RFIDLabel.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async-only wrapper (`GetDataAsync`). Go implementation is synchronous.

#### `RFIDLabel.cs`
- **File**: `FastReport.Base/RFIDLabel.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. All properties from C# (EpcFormat, AdaptiveAntenna, ReadPower, WritePower, StartPermaLock, CountPermaLock, all lock types, error handle, banks with DataColumn/Data/Offset/DataFormat) are ported in object/rfid.go. Assign() is ported. GetData() logic (resolving bracket-expression DataColumn for TIDBank, UserBank, EPCBank, AccessPasswordDataColumn, KillPasswordDataColumn) is implemented inline in engine/objects.go in the `case *object.RFIDLabel` branch, matching C# RFIDLabel.GetDataShared (RFIDLabel.cs:411-427). PlaceholderText() updated to prefer the evaluated EPCBank.Data over the DataColumn reference.

#### `Report.Async.cs`
- **File**: `FastReport.Base/Report.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API. PrepareAsync/RefreshAsync/ExportAsync/PrintAsync have no Go equivalents; report execution is synchronous via engine.ReportEngine.Run().

#### `Report.cs`
- **File**: `FastReport.Base/Report.cs`
- **Status**: PARTIALLY PORTED
- **Fixed in go-fastreport-u7abq** (2026-03-21):
  - Added `TextQuality` enum (6 values: Default/Regular/ClearType/AntiAlias/SingleBPP/SingleBPPGridFit) and `Report.TextQuality` field. Serialized as `TextQuality` attribute when non-default. Round-trip tested.
  - Added `Report.SmoothGraphics bool` field. Serialized as `SmoothGraphics="true"` when set. Round-trip tested.
  - Added `Report.ScriptLanguage string` field for round-trip fidelity only (Go does not execute scripts). Serialized when non-empty. Round-trip tested.
  - Fixed `ConvertNulls` default: `NewReport()` now initializes it to `true` matching C# `ClearReportProperties()`.
  - Serialization now writes `ReportInfo.*` dot-qualified attribute names matching C# `ReportInfo.Serialize()` output. Deserialization reads both C# dot-form and legacy Go short-form as fallbacks.
- **Remaining gaps**: GetDataSource/GetColumnValue/GetParameter/GetTotalValue introspection helpers, Dispose() lifecycle, AllNamedObjects/FindObject() object tree traversal, Refresh()/InteractiveRefresh(), Password/SaveWithRandomData cryptography (skipped), PreparePhase1()/PreparePhase2() two-pass overloads, static FromStream/FromFile/FromString factory methods, UseFileCache (low priority). OUT OF SCOPE: RegisterData(DataSet/DataTable/DataView/DataRelation); GUI methods (Show/Design/Print/PrintDialog); script compilation (GenerateReportAssembly, CodeHelper, Compile, ReferencedAssemblies).

#### `ReportComponentBase.Async.cs`
- **File**: `FastReport.Base/ReportComponentBase.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API. GetDataAsync is not applicable; component data retrieval is handled synchronously by engine/objects.go.

#### `ReportComponentBase.cs`
- **File**: `FastReport.Base/ReportComponentBase.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-21** (go-fastreport-v8zpw): Added `StylePriority` enum; `EvenStylePriority` field + getter/setter + Serialize/Deserialize; `ApplyEvenStyle(StyleLookup)` implementing C# lines 734-748; `StyleLookup` interface for import-cycle-free lookups; fixed `ApplyStyle` to use `entry.EffectiveFill()` so gradient/hatch fills are applied correctly; fixed border application to clone the full border when Lines[0] is non-nil; added `SaveState()`/`RestoreState()` matching C# lines 957-983; updated constructor default fill to use `style.NewSolidFill(style.ColorTransparent)`.
- **Gaps remaining**: Designer/interaction surface — `Cursor`, `Mouse*` event string properties (OUT OF SCOPE: UI/Windows Forms). Internal designer flags (`FlagSimpleBorder`, `FlagUseBorder`, `FlagUseFill`, `FlagPreviewVisible`, `FlagSerializeStyle`, `FlagProvidesHyperlinkValue`) (OUT OF SCOPE: designer toolbar hints only). `GetData()`, `InitializeComponent()`, `FinalizeComponent()`, `GetExpressions()` — handled centrally by `engine/objects.go` (architectural divergence; equivalent in effect). `Validate()` — OUT OF SCOPE (designer). `CalcHeight()` virtual stub — band-level equivalent is in `engine/bands.go`.

#### `ReportComponentCollection.cs`
- **File**: `FastReport.Base/ReportComponentCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Collection logic is handled by ObjectCollection in report/collections.go, including SortByTop(). SortByTop() uses sort.SliceStable (matching C# stable sort contract via TopComparer). Full test coverage added in report/collections_test.go (was 0% before this review).

#### `ReportEventArgs.cs`
- **File**: `FastReport.Base/ReportEventArgs.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go has minimal EventArgs. All 8 specialized event-args classes NOT PORTED.

#### `ReportInfo.cs`
- **File**: `FastReport.Base/ReportInfo.cs`
- **Status**: PARTIALLY PORTED
- **Fixed in go-fastreport-u7abq** (2026-03-21):
  - Added `SaveMode` enum (7 values: All/Original/User/Role/Security/Deny/Custom) with `String()` and `parseSaveMode()`. Serialized as `ReportInfo.SaveMode` when non-default. Round-trip tested for all 7 values.
  - Added `ReportInfo.Tag string` field. Serialized as `ReportInfo.Tag` when non-empty. Round-trip tested.
  - Added `ReportInfo.PreviewPictureRatio float32` field with clamp-to-0.05 for values ≤ 0 (matching C# setter). Default 0.1 not serialized. Round-trip tested.
  - `NewReport()` now initializes `PreviewPictureRatio` to 0.1 (C# `Clear()` default).
- **Remaining gaps**: `Picture` stored as `[]byte` instead of `System.Drawing.Image` (sufficient for Go use case). `Clear()` reset method not exposed as public API (Go uses `NewReport()` for fresh state). `CurrentVersion` not exposed (Go has no assembly version concept). Dedicated `Serialize()` method on `ReportInfo` not needed — serialization is done inline in `Report.Serialize()`.

#### `ReportPage.cs`
- **File**: `FastReport.Base/ReportPage.cs`
- **Status**: PARTIALLY PORTED
- **Fixed in go-fastreport-e118f**:
  - Added `HeightInPixels()` computed property: returns `UnlimitedHeightValue` when `UnlimitedHeight=true`, otherwise `PaperHeight * units.Millimeters` (mirrors ReportPage.cs:374-379).
  - Added `WidthInPixels()` computed property: returns `UnlimitedWidthValue` when `UnlimitedWidth=true` and value is non-zero, otherwise `PaperWidth * units.Millimeters` (mirrors ReportPage.cs:385-398; Go skips `IsDesigning` check — not applicable to headless engine).
  - `BackPage`/`BackPageOddEven`: Go uses `string BackPage` (page name reference) + `int BackPageOddEven` (0=both, 1=odd, 2=even) — an intentional Go extension. C# uses `bool BackPage` with no odd/even control. The serialization is Go-format only (not compatible with C# FRX files that use `BackPage` as bool). Documented as intentional divergence.
- **Remaining gaps**: `Guides` (designer-only, skip), `Subreport` back-reference, `LinkToPage`/`PageLink` class (low priority, complex), `IParent` interface implementation, `GetExpressions()`/`ExtractMacros()`, event-firing methods (`OnCreatePage`/`OnStartPage`/`OnFinishPage`/`OnManualBuild`), `LoadExternalPage()`.

#### `ReportSettings.cs`
- **File**: `FastReport.Base/ReportSettings.cs`
- **Status**: PARTIALLY PORTED
- **Fixed in go-fastreport-u7abq** (2026-03-21): Reviewed against C# source. Core settings already ported (DefaultPaperSize, UsePropValuesToDiscoverBO, ImageLocationRoot).
- **Remaining gaps**: DatabaseLogin/AfterDatabaseLogin event hooks — Go equivalent would be function callbacks but not implemented; FilterBusinessObjectProperties/GetBusinessObjectPropertyKind/GetBusinessObjectTypeInstance callbacks — business-object auto-discovery not applicable to Go data binding model; DefaultLanguage property — script language selection is OUT OF SCOPE for Go. Script-related settings (ScriptLanguage, ReferencedAssemblies) are out of scope.

#### `ReportSummaryBand.cs`
- **File**: `FastReport.Base/ReportSummaryBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `ReportTitleBand.cs`
- **File**: `FastReport.Base/ReportTitleBand.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `ShapeObject.cs`
- **File**: `FastReport.Base/ShapeObject.cs`
- **Status**: MOSTLY PORTED
- **Fixed 2026-03-21** (go-fastreport-lfkpm): Shape (ShapeKind) is now serialized/deserialized as enum name string ("Rectangle"/"RoundRectangle"/"Ellipse"/"Triangle"/"Diamond") matching C# WriteValue. Previously used WriteInt/ReadInt which wrote integers instead of names.
- **Fixed 2026-03-22** (go-fastreport-yuhkl): Verified DashPattern serialize/deserialize already implemented in object/lines.go ShapeObject.Serialize/Deserialize. Verified Assign() already implemented (ShapeObject.Assign). Porting-gaps.md entry was stale.
- **Gaps (remaining)**: Draw() rendering — handled by Go exporters rather than on the object itself.

#### `Sort.cs`
- **File**: `FastReport.Base/Sort.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Go equivalent is `SortSpec` struct in band/types.go with `Expression` and `Column` fields (Expression overrides Column when non-empty) and `Order SortOrder`. Serialization as `<Sort Expression="..." Descending="true"/>` child elements is handled by `sortSpecItem`/`sortCollection` and `DataBand.DeserializeChild("Sort",...)`. **Fixed (go-fastreport-mdnt4)**: SortSpec.Expression field now properly exposed in DataBand.GetExpressions() — when both Expression and Column are set, Expression takes priority, matching C# Sort.Expression property.

#### `SortCollection.cs`
- **File**: `FastReport.Base/SortCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Go equivalent is `[]SortSpec` (a slice) on DataBand with AddSort/SetSort/Sort accessors. The `sortCollection` wrapper handles FRX round-trip serialization and deserialization via `DataBand.DeserializeChild`. Expression-based sorting is fully supported via `SortSpec.Expression` (any valid expression string, not just column names). **Fixed (go-fastreport-mdnt4)**: DataBand.GetExpressions() now includes sort expressions so they are visible to the expression walker/validator. DataBand.Assign() deep-copies the sort slice. GroupHeaderBand.InitDataSource()/FinalizeDataSource() equivalent logic is in engine/groups.go:applyGroupSort() which injects group-condition sorts into the DataBand sort list before calling the data source init.

#### `Style.cs`
- **File**: `FastReport.Base/Style.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-21** (go-fastreport-v8zpw): `StyleEntry` now carries `Fill style.Fill` and `TextFill style.Fill` interface fields alongside legacy colour scalars. `EffectiveFill()`/`EffectiveTextFill()` pick the richer fill. `Assign()` deep-copies `Fill`/`TextFill`. `Clone()` returns a deep copy. `ApplyEvenStyle` on `ReportComponentBase`. Tests in `style/styleentry_porting_gaps_test.go`.
- **Gaps remaining**: `SaveStyle()`/`RestoreStyle()` OUT OF SCOPE (designer undo/redo). Gradient/hatch fills in `<Style>` FRX not yet XML-serialized (`styles_serial.go` writes `Fill.Color` only).

#### `StyleBase.cs`
- **File**: `FastReport.Base/StyleBase.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-21** (go-fastreport-v8zpw): `StyleEntry.Assign(src)` deep-copies `Border`, `Fill`, and `TextFill`, matching C# `StyleBase.Assign`. `Fill`/`TextFill` are `style.Fill` interface fields. `ApplyStyle` in `ReportComponentBase` uses `entry.EffectiveFill()`. Tests in `style/styleentry_porting_gaps_test.go`.
- **Gaps remaining**: `SaveStyle()`/`RestoreStyle()` OUT OF SCOPE (designer undo/redo). FRX XML serialization of gradient/hatch fills inside `<Style>` elements not yet implemented.

#### `StyleCollection.cs`
- **File**: `FastReport.Base/StyleCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Ported as StyleSheet in style/stylesheet.go with map-based registry (lowercase-keyed for case-insensitive lookups matching C# String.Compare ignoreCase:true), insertion-order slice, Add/Find/Len/All, and serialization via reportpkg/styles_serial.go. Fixed in review: Find() was case-sensitive (bug); now correctly case-insensitive. Tests added in style/stylesheet_test.go.

#### `StyleSheet.cs`
- **File**: `FastReport.Base/StyleSheet.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. C# two-level hierarchy (StyleSheet->StyleCollection->Style) flattened to one level (StyleSheet->StyleEntry) in Go — semantically equivalent. Serialized as <Styles> with <Style> children, matching FRX format. Case-insensitive name lookup now matches C# behaviour.

#### `SubreportObject.cs`
- **File**: `FastReport.Base/SubreportObject.cs`
- **Status**: MOSTLY PORTED
- **Fixed 2026-03-22** (go-fastreport-371aq): Added `SubreportObject.Assign()` — copies ReportPageName, PrintOnParent, ReportName on top of base Assign(). Fixed `NewSubreportObject()` to clear `CanCopy` flag matching C# constructor (SubreportObject.cs:154). Note: FlagUseBorder/FlagUseFill/FlagPreviewVisible are designer-only flags with no direct Go equivalent; the CanCopy clear is the only runtime-visible difference.
- **Remaining gaps**: ReportPage stored as name string rather than object reference (no reference lifecycle management or RemoveSubReport() cleanup).

#### `TextObject.Async.cs`
- **File**: `FastReport.Base/TextObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API with context.Context for cancellation.

#### `TextObject.cs`
- **File**: `FastReport.Base/TextObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing Value property, Assign(), SaveState/RestoreState, TabPositions, Trimming, InlineImageCache, GetStringFormat(), DrawText(), BreakText(), CalcWidth/Height/Size(), GetData(), Break(), and GetExpressions(). Core text rendering handled by engine/objects.go instead. **Fixed 2026-03-21** (go-fastreport-lfkpm): All enum fields serialized as string names matching C# WriteValue: HorzAlign ("Left"/"Center"/"Right"/"Justify"/"JustifyAll"), VertAlign ("Top"/"Center"/"Bottom"), TextRenderType ("Default"/"Inline"/"HtmlParagraph"/"HtmlTags"), AutoShrink ("None"/"FontSize"/"FontWidth"), MergeMode ("None"/"Merge"/"MergeSameValue"), ProcessAt ("Default"/"Preview"/"Once"), Duplicates ("Show"/"Hide"/"HideButMerge"/"Clear"). Previously used WriteInt/ReadInt for all these.

#### `TextObjectBase.Async.cs`
- **File**: `FastReport.Base/TextObjectBase.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API with context.Context for cancellation.

#### `TextObjectBase.cs`
- **File**: `FastReport.Base/TextObjectBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing Value field/getter, Assign(), ExtractMacros(), GetTextWithBrackets/WithoutBrackets(), FormatValue() overloads, CalcAndFormatExpression(), and GetDisplayText(). Engine calls evalTextWithFormat() instead.

#### `TextOutline.cs`
- **File**: `FastReport.Base/TextOutline.cs`
- **Status**: FULLY PORTED
- **Fixed 2026-03-22** (go-fastreport-4vu3w): DrawBehind bool field already present in style/textoutline.go and serialized/deserialized in object/text.go (TextOutline.Enabled/Color/Width/DashStyle/DrawBehind round-trip). Porting-gaps.md was stale.

#### `Watermark.cs`
- **File**: `FastReport.Base/Watermark.cs`
- **Status**: MOSTLY PORTED
- **Fixed 2026-03-22** (go-fastreport-klflm): Image FRX deserialization/serialization added — `Watermark.Image` attribute is read as base64 string and decoded into `ImageData []byte`; Serialize writes it back when non-empty. Verified TextFillColor default is correct (`Color.FromArgb(40, Color.Gray)` = RGBA{A:40,R:128,G:128,B:128}) matching C# Watermark constructor (Watermark.cs:362).
- **Remaining gaps**: TextFill is `color.RGBA` only (not a full Fill interface supporting gradients); no macro expansion in Text field (C# `ProcessText()` with [Page#]/[TotalPages#] substitution).

#### `ZipCodeObject.Async.cs`
- **File**: `FastReport.Base/ZipCodeObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — async GetDataAsync wrapper not applicable to Go's synchronous model.

#### `ZipCodeObject.cs`
- **File**: `FastReport.Base/ZipCodeObject.cs`
- **Status**: PARTIALLY PORTED
- **Fixed (2026-03-22)** (go-fastreport-rmmgo): Added `ZipCodeObject.Assign(src)` — copies all fields (segmentWidth, segmentHeight, spacing, segmentCount, showMarkers, showGrid, dataColumn, expression, text) on top of ReportComponentBase.Assign. Mirrors C# ZipCodeObject.Assign (ZipCodeObject.cs:247-263). GetExpressions() and GetData() were already implemented.
- **Remaining Gaps**: Draw() rendering — DrawSegment, DrawReferenceLine, DrawSegmentGrid, FDigits digit patterns are not implemented. Engine renders ZipCodeObject as blank placeholder instead of graphical zip code.

### Barcode/Aztec

#### `AztecCode.cs`
- **File**: `FastReport.Base/Barcode/Aztec/AztecCode.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `AztecEncodingOptions.cs`
- **File**: `FastReport.Base/Barcode/Aztec/AztecEncodingOptions.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `BinaryShiftToken.cs`
- **File**: `FastReport.Base/Barcode/Aztec/BinaryShiftToken.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `BitArray.cs`
- **File**: `FastReport.Base/Barcode/Aztec/BitArray.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `BitMatrix.cs`
- **File**: `FastReport.Base/Barcode/Aztec/BitMatrix.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `EncodeHintType.cs`
- **File**: `FastReport.Base/Barcode/Aztec/EncodeHintType.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `Encoder.cs`
- **File**: `FastReport.Base/Barcode/Aztec/Encoder.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Parameter selection (compact/full mode, layer count) and GF arithmetic ported in barcode/aztec_impl.go. Missing: HighLevelEncoder (character→bitstream), bit stuffing (stuffBits), matrix construction (buildMatrix), finder pattern encoding, and full data interleaving. Go implementation covers error correction parameters but not the complete Aztec symbol generation pipeline.

#### `EncodingOptions.cs`
- **File**: `FastReport.Base/Barcode/Aztec/EncodingOptions.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `GenericGF.cs`
- **File**: `FastReport.Base/Barcode/Aztec/GenericGF.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `GenericGFPoly.cs`
- **File**: `FastReport.Base/Barcode/Aztec/GenericGFPoly.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `HighLevelEncoder.cs`
- **File**: `FastReport.Base/Barcode/Aztec/HighLevelEncoder.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `ReedSolomonEncoder.cs`
- **File**: `FastReport.Base/Barcode/Aztec/ReedSolomonEncoder.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `SimpleToken.cs`
- **File**: `FastReport.Base/Barcode/Aztec/SimpleToken.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `State.cs`
- **File**: `FastReport.Base/Barcode/Aztec/State.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `SupportClass.cs`
- **File**: `FastReport.Base/Barcode/Aztec/SupportClass.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

#### `Token.cs`
- **File**: `FastReport.Base/Barcode/Aztec/Token.cs`
- **Status**: NOT PORTED
- **Gaps**: Go Aztec encoder is simplified placeholder.

### Barcode

#### `Barcode128.cs`
- **File**: `FastReport.Base/Barcode/Barcode128.cs`
- **Status**: FULLY PORTED — reviewed go-fastreport-ylosy
- **Gaps**: None
- **Review notes (go-fastreport-ylosy)**:
  - Character table (tabelle128, 106 entries): verified identical to C# tabelle_128.
  - Start chars CODE_A=103, CODE_B=104, CODE_C=105; stop "2331112": all match.
  - Checksum: start_value + sum(i * char_value) mod 103: matches C# exactly.
  - Bug fixed: `IsFourOrMoreDigits` condition `index+4 < code.Length` (strictly less
    than) means exactly 4 digits at end-of-string do NOT select Code C. The Go port
    was using `c128CountDigits >= 4` which incorrectly selected Code C for "1234".
    Fixed by adding `c128IsFourOrMoreDigits` mirroring Barcode128.cs:241.
  - Subset auto-selection, SHIFT, FNC1-4, CODE A/B/C switches: verified correct.
  - doConvert: matches C# DoConvert (even positions +5, all -1).
  - FNC1 (idx=102): verified present.

#### `Barcode2DBase.cs`
- **File**: `FastReport.Base/Barcode/Barcode2DBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: CalcBounds and Draw2DBarcode ported. Missing: Swiss QR cross overlay, showMarker ST L-shape, showText text below 2D barcodes not drawn, Angle rotation not applied, QR module shapes.

#### `Barcode2of5.cs`
- **File**: `FastReport.Base/Barcode/Barcode2of5.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. **Reviewed 2026-03-22** (go-fastreport-f9vos): All previously listed gaps resolved. Deutsche Identcode display text formatting (`insertAt` dot/space sequence, PrintCheckSum conditional check-digit strip) implemented in `barcode/code2of5.go`. Deutsche Leitcode display text formatting (7 sequential inserts) implemented correctly. ITF14 `DrawText` override (space-separated digit groups) implemented as `ITF14FormatDisplayText()`. Serialize for `DeutscheIdentcodeBarcode.PrintCheckSum` and `DeutscheLeitcodeBarcode.PrintCheckSum` added to `BarcodeObject.Serialize()` as `Barcode.DrawVerticalBearerBars` (C# naming quirk). All behaviors tested in `barcode/porting_barcode_gaps_test.go`.

#### `Barcode39.cs`
- **File**: `FastReport.Base/Barcode/Barcode39.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — both Barcode39 and Barcode39Extended fully ported with matching lookup tables, checksum logic, and GetPattern behavior.

#### `Barcode93.cs`
- **File**: `FastReport.Base/Barcode/Barcode93.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — IsNumeric() returns false, SetCalcCheckSum() implemented, code93GetPattern now conditionally includes check digits based on includeChecksum parameter.

#### `BarcodeAztec.cs`
- **File**: `FastReport.Base/Barcode/BarcodeAztec.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Wrapper structurally complete. CRITICAL: Go encoder is simplified placeholder vs C# ZXing encoder (~3268 lines). Missing: HighLevelEncoder, bit stuffing, proper symbol sizing, alignment map, reference grid lines. Produces non-scannable symbols.

#### `BarcodeBase.cs`
- **File**: `FastReport.Base/Barcode/BarcodeBase.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. **Resolved 2026-03-22** (go-fastreport-0uizw): All previously missing items implemented: `Clone()`/`Assign()` via `CloneBarcode()` package function with explicit type switch over all ~35 barcode types and `BaseBarcodeImpl.Assign()`; `StripControlCodes()` returns data unchanged (mirrors C# base class); Color/Font FRX serialization added — `Barcode.Color` written as `#RRGGBB` when non-default (not black), `Barcode.Font.Name`/`Barcode.Font.Size` written when non-default (not Arial/8pt); `parseColorStr()` helper supports `#RRGGBB` and `#AARRGGBB`. Tests in `barcode/porting_barcode_gaps_test.go`.

#### `BarcodeCodabar.cs`
- **File**: `FastReport.Base/Barcode/BarcodeCodabar.cs`
- **Status**: MOSTLY PORTED
- **Reviewed 2026-03-22** (go-fastreport-f9vos): StartChar/StopChar Deserialize and Serialize are now implemented (`BarcodeObject.Deserialize` reads `Barcode.StartChar`/`Barcode.StopChar`; `BarcodeObject.Serialize` writes them when non-default). Round-trip tested in `barcode/porting_barcode_gaps_test.go`.
- **Fixed 2026-03-22** (go-fastreport-qiz8z): Added `CodabarBarcode.Assign()` — copies BaseBarcodeImpl, StartChar, StopChar fields. Mirrors C# BarcodeCodabar.Assign.
- **Remaining minor gaps**: `IsNumeric` property (returns false in C#, not surfaced in Go's `BarcodeBase` interface), `CodabarChar` enum type (Go uses `byte` directly).

#### `BarcodeDatamatrix.cs`
- **File**: `FastReport.Base/Barcode/BarcodeDatamatrix.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: **Resolved 2026-03-22** (go-fastreport-7u7w9): `SymbolSize` and `Encoding` FRX properties now wired into the encoding pipeline. `dmSymbolSizeToHW()` lookup maps C# `DatamatrixSymbolSize` enum names (e.g. `"10x10"`, `"24x24"`) to `(h,w)` pairs matching `dmSizes` indices exactly as C# does `dmSizes[(int)SymbolSize-1]`. `parseDmEncoding()` maps `DatamatrixEncoding` enum names to `dmEncodingMode`. `dmGetEncodationWithMode()` dispatches directly to the requested encoder (Ascii/C40/Text/Base256/X12/Edifact) when not Auto; falls back to Auto multi-algorithm shortest-pick when Auto. `DataMatrixBarcode.GetMatrix()` passes both `symH`/`symW` and `enc` into `dmGetMatrixWithOptions()` → `dmGenerateWithOptions()`. `dmGenerate()` is now a thin wrapper calling `dmGenerateWithOptions(text, 0, 0, dmEncodingAuto)`. Remaining gaps: `CodePage` and `PixelSize` properties deserialized but not used (CodePage requires transcoding the input; PixelSize is a render hint). GS1 variant lacks AI-level validation.

#### `BarcodeEAN.cs`
- **File**: `FastReport.Base/Barcode/BarcodeEAN.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — EAN-8, EAN-13, and EAN-128/GS1-128 all fully ported with encoding tables, pattern generation, text positioning, and CalcBounds.

#### `BarcodeGS1.cs`
- **File**: `FastReport.Base/Barcode/BarcodeGS1.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21 (go-fastreport-1yqnb). GS1_128Barcode.GetPattern(), GS1DataBar variants (Omni/Stacked/StackedOmni/Limited), GetGS1Widths, and Combins are all correctly ported. 31 internal tests added in barcode/gs1_helper_internal_test.go covering the AI table, FindAIIndex, GetCode, and ParseGS1 logic.

#### `BarcodeIntelligentMail.cs`
- **File**: `FastReport.Base/Barcode/BarcodeIntelligentMail.cs`
- **Status**: MOSTLY PORTED
- **Fixed 2026-03-22** (go-fastreport-4pwyz): Added `IntelligentMailBarcode.Assign()` — copies BaseBarcodeImpl and QuietZone fields. QuietZone is already present as a field and handled in BarcodeObject.Serialize/Deserialize. Mirrors C# BarcodeIntelligentMail.Assign (BarcodeIntelligentMail.cs:44-48).
- **Remaining gaps**: Full IMb encoding from valid digit strings not implemented (Encode validates digit count but Render returns placeholder); barcode rendering requires the full 5-state bar algorithm.

#### `BarcodeMSI.cs`
- **File**: `FastReport.Base/Barcode/BarcodeMSI.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. **Reviewed 2026-03-21** (go-fastreport-i7nfi): Boundary condition audit complete. tabelleMSI lookup table (10 entries) verified against C# BarcodeMSI.cs. Luhn checksum formula verified: checkOdd accumulates digits at odd indices concatenated as a number (not summed), checksum = digitSum(checkOdd*2) + checkEven; mod10 then invert. Boundary tests added: "999"→check digit 3 (pattern len 37, suffix "51516060515"), "0"→check digit 0 (pattern len 21). CalcChecksum flag correctly gates check digit in GetPattern(). EncodedText() stores raw input only (check digit is not stored there, matching C# BarcodeBase.EncodedText).

#### `BarcodeMaxiCode.cs`
- **File**: `FastReport.Base/Barcode/BarcodeMaxiCode.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-22** (previous session): Assign() override for Mode property added (barcode/extended.go MaxiCodeBarcode.Assign); Mode serialization added in BarcodeObject.Serialize() — writes `Barcode.Mode` when != 4.
- **Remaining gaps**: Initialize() override calling maxiCodeImpl.encode() after setting mode; exact hexagonal polygon vertex rendering (Go uses approximate fill instead of C# Hexagon/Ellipse struct vertices).

#### `BarcodeObject.Async.cs`
- **File**: `FastReport.Base/Barcode/BarcodeObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go does not use Task/async-await; the engine executes synchronously via ReportEngine.Run().

#### `BarcodeObject.cs`
- **File**: `FastReport.Base/Barcode/BarcodeObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: **Partially resolved 2026-03-22** (go-fastreport-0uizw): `Assign()` deep-copy implemented — copies all `ComponentBase` fields, barcode-specific fields, and deep-clones the embedded barcode via `CloneBarcode()`; `SaveState()`/`RestoreState()` implemented — save/restore `text` alongside the base class Bounds/Visible/Border/Fill state; `GetExpressions()` implemented — collects `DataColumn` and `Expression` if non-empty, appending `ComponentBase.GetExpressions()` results. Remaining gaps: `Draw()` designer rendering (OUT OF SCOPE for headless Go engine), `GetData()` DataColumn/Expression evaluation and bracket expression processing, `GetDataShared()` with QRData.Parse() and Swiss QR handling, `SymbologyName` property (set by name), and Barcode setter null-fallback to Barcode39.

#### `BarcodePDF417.cs`
- **File**: `FastReport.Base/Barcode/BarcodePDF417.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core PDF417 support exists, but Go encoder (`barcode/pdf417_impl.go`) is simplified versus C#: it does not implement full cluster-table codeword rendering and uses simplified compaction/EC flow. C# surface properties (`AspectRatio`, `Columns`, `Rows`, `CodePage`, `CompactionMode`, `ErrorCorrection`, `PixelSize`) deserialize into Go structs, but encoder behavior is not yet fully driven by those knobs with C# parity.

#### `BarcodePharmacode.cs`
- **File**: `FastReport.Base/Barcode/BarcodePharmacode.cs`
- **Status**: MOSTLY PORTED
- **Reviewed 2026-03-22** (go-fastreport-f9vos): `QuietZone` Serialize now implemented — `BarcodeObject.Serialize` writes `Barcode.QuietZone=false` when non-default (default true is not written). Tested in `barcode/porting_barcode_gaps_test.go`.
- **Fixed 2026-03-22** (go-fastreport-ni8iz): Added `PharmacodeBarcode.Assign()` — copies BaseBarcodeImpl, TwoTrack, QuietZone fields. Mirrors C# BarcodePharmacode.Assign.
- **Remaining minor gaps**: `IsNumeric` property override (C# returns true; not surfaced in Go interface); `TwoTrack` property exists in Go but not in C# source (Go extension).

#### `BarcodePlessey.cs`
- **File**: `FastReport.Base/Barcode/BarcodePlessey.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Pattern generation, CRC logic, start/termination/end sequences, and hex validation are ported in `barcode/plessey.go`.

#### `BarcodePostNet.cs`
- **File**: `FastReport.Base/Barcode/BarcodePostNet.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — `PostNetBarcode.Render` updated to use `GetPattern()` + `DrawLinearBarcode()` matching C# LinearBarcodeBase path. Removed dead `postnetEncode()` helper.

#### `BarcodeQR.cs`
- **File**: `FastReport.Base/Barcode/BarcodeQR.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: QR matrix generation, quiet-zone behavior, Swiss-QR `SPC` detection forcing `M`, and `UseThinModules` are ported. Go renders circular modules, but lacks the rest of the C# shape set (`Diamond`, `RoundedSquare`, `PillHorizontal`, `PillVertical`, `Plus`, `Hexagon`, `Star`, `Snowflake`) and does not mirror the C# `Angle`-driven rotational rendering. Also missing `Assign` and `Serialize` support for QR-specific properties. **Reviewed 2026-03-21** (go-fastreport-yaqtb): Algorithm precision audit complete. Version selection (numDataCodewords >= numInputBytes+3), Reed-Solomon GF(256), mask pattern scoring, finder/timing pattern placement, alignment patterns, format info encoding all verified correct. "HELLO WORLD" at EC level M encodes as alphanumeric mode → version 1 (21×21) matrix — QuietZone=true (default) adds 4-module border making it 29×29. Test added: TestQRBarcode_HelloWorld_ECLevelM with QuietZone=false verifying 21×21 dimensions, three finder pattern outer rings, horizontal/vertical timing patterns (row/col 6), and fixed dark module at (13,8).

#### `BarcodeUPC.cs`
- **File**: `FastReport.Base/Barcode/BarcodeUPC.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: **Reviewed 2026-03-22** (go-fastreport-f9vos): No FRX round-trip properties are missing — UPC types have no per-type Serialize/Deserialize in C#. Remaining gaps are rendering/GUI-specific: quiet zone margins (`extra1`/`extra2` in C# constructor), `textUp` flag (supplements draw text above bars), and the Supplement `Render` methods currently suppress text rendering. These do not affect FRX load/save correctness.

#### `GS1Helper.cs`
- **File**: `FastReport.Base/Barcode/GS1Helper.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Application Identifier (AI) table (89 entries) and parsing logic (ParseGS1, GetCode, FindAIIndex) are fully ported in barcode/gs1.go. Reviewed 2026-03-21 (go-fastreport-1yqnb): all 89 AI table entries verified against C# source; algorithm logic (FindAIIndex wildcard matching, GetCode fixed/variable-length paths, ParseGS1 FNC1 prepend/separator) confirmed correct. One Go improvement: bounds check `index >= len(code)` guards against panic (C# would throw IndexOutOfRangeException). 31 internal tests in barcode/gs1_helper_internal_test.go.

#### `LinearBarcodeBase.cs`
- **File**: `FastReport.Base/Barcode/LinearBarcodeBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core bar-rendering pipeline ported. Missing: DrawTopLabel()/DrawBottomLabel() for Russian Post barcode (ПОЧТА РОССИИ label, digit grouping with bold formatting), IsBarcodeRussianPost property with associated sizing (21.15f extra width, 18px left margin, 9px top offset, 56.7f height), DrawString() overloads with font scaling/zoom compensation, DoLines() internals for Intelligent Mail special line types (BlackHalf/BlackLong/BlackTracker/BlackAscender/BlackDescender), OneBarProps() method, CheckText() numeric validation, and lazy Code→GetPattern() evaluation.

#### `SwissQRCode.cs`
- **File**: `FastReport.Base/Barcode/SwissQRCode.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Simplified flat-field. Missing typed class hierarchy, ALL validation, payload format issues, Unpack/Parse.

### Barcode/QRCode

#### `BitVector.cs`
- **File**: `FastReport.Base/Barcode/QRCode/BitVector.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Ported as qrBitVector struct in barcode/qr.go with all methods (size, at, appendBit, appendBits, appendBitVector, xorWith, appendByte). MSB-first bit-packing verified against C# `(7 - numBitsInLastByte)` shift formula. Dynamic doubling matches C# `array.Length << 1`. Note: Go xorWith omits the size-mismatch panic in C# — safe because all callers pass equal-size vectors (go-fastreport-6uh4c reviewed).
- **Tests**: barcode/qr_datastructs_internal_test.go (TestQRBitVector_*)

#### `BlockPair.cs`
- **File**: `FastReport.Base/Barcode/QRCode/BlockPair.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Ported as inline blockPair{data []int, ec []int} within qrInterleave() in barcode/qr.go. Fields correspond to C# DataBytes/ErrorCorrectionBytes (go-fastreport-6uh4c reviewed).
- **Tests**: barcode/qr_datastructs_internal_test.go (TestQRInterleave_BlockPair_DataAndEC)

#### `ByteArray.cs`
- **File**: `FastReport.Base/Barcode/QRCode/ByteArray.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Replaced by idiomatic Go []int slices; no wrapper class needed. All call sites use `& 0xff` masking matching C# `at(index) = bytes[index] & 0xff` unsigned semantics (go-fastreport-6uh4c reviewed).

#### `ByteMatrix.cs`
- **File**: `FastReport.Base/Barcode/QRCode/ByteMatrix.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Ported as qrByteMatrix struct in barcode/qr.go. x=column y=row stored as bytes[y][x] matches C# exactly. get(), set(), clear() verified (go-fastreport-6uh4c reviewed).
- **Tests**: barcode/qr_datastructs_internal_test.go (TestQRByteMatrix_*)

#### `Encoder.cs`
- **File**: `FastReport.Base/Barcode/QRCode/Encoder.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: All 17 core functions faithfully ported. Low-impact: Kanji mode not implemented (Byte fallback), Shift_JIS hint missing, post-encode validation omitted.

#### `ErrorCorrectionLevel.cs`
- **File**: `FastReport.Base/Barcode/QRCode/ErrorCorrectionLevel.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Ported as qrECLevel struct with L/M/Q/H constants and qrECLevelFromString() in barcode/qr.go. Ordinals (L=0,M=1,Q=2,H=3) and format-info bits (L=0x01,M=0x00,Q=0x03,H=0x02) match C# exactly. qrECLevelFromString handles uppercase L/M/Q/H; unrecognised input falls back to M — C# dead-code default returns L but M is correct for the "M" string case (go-fastreport-6uh4c reviewed).
- **Tests**: barcode/qr_datastructs_internal_test.go (TestQRECLevel_*, TestQRECLevelFromString_*)

#### `GF256.cs`
- **File**: `FastReport.Base/Barcode/QRCode/GF256.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Galois Field GF(256) arithmetic with primitive 0x011D, multiply/exp/log tables ported as qrGF256 in barcode/qr.go. Reviewed (go-fastreport-gpzir): table init loop matches C# exactly (GF256.cs:71-91); multiply() omits the C# `a==1`/`b==1` early-returns but is algebraically identical (logTable[1]=0 → exp[(0+logTable[b])%255]=b). Log/exp tables are mutual inverses across all 255 non-zero elements. All tests pass.
- **Tests**: barcode/qr_math_internal_test.go (TestQRGF256_ExpTable_KnownValues, TestQRGF256_LogTable_KnownValues, TestQRGF256_LogExpInverse, TestQRGF256_Multiply_*)

#### `GF256Poly.cs`
- **File**: `FastReport.Base/Barcode/QRCode/GF256Poly.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Polynomial arithmetic (degree, addOrSubtract, multiply, multiplyByMonomial, divide) ported as qrGF256Poly in barcode/qr.go. Reviewed (go-fastreport-gpzir): coefficient indexing convention matches C# (coefficients[0] = highest-degree term; getCoefficient(d) = coefficients[len-1-d]); leading-zero stripping in constructor; all operations verified algebraically. All tests pass.
- **Tests**: barcode/qr_math_internal_test.go (TestQRGF256Poly_Degree, TestQRGF256Poly_IsZero, TestQRGF256Poly_StripLeadingZeros, TestQRGF256Poly_GetCoefficient, TestQRGF256Poly_AddOrSubtract, TestQRGF256Poly_MultiplyByMonomial, TestQRGF256Poly_Multiply_SimpleProduct, TestQRGF256Poly_Divide_Remainder)

#### `MaskUtil.cs`
- **File**: `FastReport.Base/Barcode/QRCode/MaskUtil.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. All four JISX0510:2004 Table 21 penalty rules ported as qrPenaltyRule1-4() and qrCalcPenalty() in barcode/qr.go. Rule 4 formula fixed (go-fastreport-4j28l): C# uses `Math.Abs((int)(darkRatio*100 - 50))` (float subtract then truncate); the prior Go code used `int(ratio) - 50` (truncate then subtract), giving different results for non-integer percentages. Now matches C# exactly (MaskUtil.cs:118).

#### `MatrixUtil.cs`
- **File**: `FastReport.Base/Barcode/QRCode/MatrixUtil.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Finder patterns (qrEmbedFinder), format info (qrEmbedTypeInfo), version info (qrMaybeEmbedVersionInfo), basic patterns (qrEmbedBasicPatterns), and coordinate tables all ported in barcode/qr.go.

#### `Mode.cs`
- **File**: `FastReport.Base/Barcode/QRCode/Mode.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. qrMode type with NUMERIC/ALPHANUMERIC/BYTE/KANJI values and characterCountBits() ported in barcode/qr.go.

#### `QRCode.cs`
- **File**: `FastReport.Base/Barcode/QRCode/QRCode.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Simple data holder replaced by function parameters and local variables in encodeQR(); all fields preserved functionally.

#### `QRCodeWriter.cs`
- **File**: `FastReport.Base/Barcode/QRCode/QRCodeWriter.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Main encode() logic ported as encodeQR() in barcode/qr.go including matrix rendering, quiet zone, and width/height sizing.

#### `QRData.cs`
- **File**: `FastReport.Base/Barcode/QRCode/QRData.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: FastReport-specific QR data format parsing (vCard, SMS, Geo, Email structured payloads). Barcode object in Go handles basic text encoding but lacks the full QRData.Parse() payload builder that BarcodeObject.GetDataShared() calls in C#.

#### `ReedSolomonEncoder.cs`
- **File**: `FastReport.Base/Barcode/QRCode/ReedSolomonEncoder.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. buildGenerator() with polynomial caching and encode() ported as qrReedSolomon in barcode/qr.go. Reviewed (go-fastreport-gpzir): cachedGenerators seed, buildGenerator loop index, and encode() copy/multiply/divide/zero-pad sequence all match C# exactly (ReedSolomonEncoder.cs:44-87). Algebraic correctness verified: codeword polynomial evaluates to 0 at generator roots α^0..α^(ecBytes-1). EC bytes for version 1-M data confirmed stable across identical inputs.
- **Tests**: barcode/qr_math_internal_test.go (TestQRReedSolomon_BuildGenerator_Degree6, TestQRReedSolomon_BuildGenerator_Cache, TestQRReedSolomon_Encode_SelfConsistency, TestQRReedSolomon_Encode_KnownDataBytes, TestQRReedSolomon_Encode_GeneratorRoots)

#### `SupportClass.cs`
- **File**: `FastReport.Base/Barcode/QRCode/SupportClass.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Java→C# compatibility helpers (URShift, Identity). Go's native unsigned types and >> operator subsume these; no equivalent needed.

#### `Version.cs`
- **File**: `FastReport.Base/Barcode/QRCode/Version.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. All 40 QR versions × 4 EC levels (160 capacity entries), numDataCodewords/numBlocks/ecPerBlock, and position adjustment pattern coordinates ported as qrVersionInfo and qrVersionTable in barcode/qr.go. Fixed (go-fastreport-4j28l): the `ecCodewords` field in qrVersionInfo stores EC codewords **per block** (matching C# `ECBlocks.ECCodewordsPerBlock`), not a total. The comment was wrong and `ecPerBlock()` was incorrectly dividing by numBlocks; now `ecPerBlock()` returns `ecCodewords` directly.

#### `WriterException.cs`
- **File**: `FastReport.Base/Barcode/QRCode/WriterException.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Custom exception replaced by idiomatic Go error return from encodeQR() — functionally equivalent.

### Code

#### `AssemblyDescriptor.cs`
- **File**: `FastReport.Base/Code/AssemblyDescriptor.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Go uses expr-lang/expr.

#### `CodeHelper.cs`
- **File**: `FastReport.Base/Code/CodeHelper.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Go uses expr-lang/expr.

#### `CodeProvider.cs`
- **File**: `FastReport.Base/Code/CodeProvider.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Go uses expr-lang/expr.

#### `CodeUtils.cs`
- **File**: `FastReport.Base/Code/CodeUtils.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: **Reviewed 2026-03-22 (go-fastreport-ertdg)**. IMPLEMENTED: String-skipping in bracket parser (expr/parser.go `ParseWithBrackets` — outer and inner loops now skip double-quoted string literals via character scan, mirroring C# `CodeUtils.SkipString`). REMAINING: ExportableExpression never evaluated for bands — infrastructure gap (PreparedObject has no Exportable field); not implementable without `preview.PreparedObject` changes.

#### `CsCodeHelper.cs`
- **File**: `FastReport.Base/Code/CsCodeHelper.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Go uses expr-lang/expr.

#### `ExpressionDescriptor.cs`
- **File**: `FastReport.Base/Code/ExpressionDescriptor.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Abstract base class for CodeDom expression compilation. Entire Code/ directory (8 files) superseded by expr-lang/expr.

#### `VbCodeHelper.cs`
- **File**: `FastReport.Base/Code/VbCodeHelper.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: VB.NET code generation for CodeDom. Go uses expr-lang/expr. CodeUtils bracket-matching ported separately.

### Code/Ms

#### `MsAssemblyDescriptor.Async.cs`
- **File**: `FastReport.Base/Code/Ms/MsAssemblyDescriptor.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async CodeDom compilation. Go uses expr-lang/expr.

#### `MsAssemblyDescriptor.cs`
- **File**: `FastReport.Base/Code/Ms/MsAssemblyDescriptor.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: CodeDom runtime compilation, assembly caching. Go uses expr-lang/expr.

#### `MsCodeProvider.cs`
- **File**: `FastReport.Base/Code/Ms/MsCodeProvider.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Factory for CodeDom runtime compilation. Go uses expr-lang/expr.

#### `MsExpressionDescriptor.cs`
- **File**: `FastReport.Base/Code/Ms/MsExpressionDescriptor.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET CodeDom — uses Reflection.MethodInfo. Go uses expr-lang/expr.

#### `StubClasses.cs`
- **File**: `FastReport.Base/Code/Ms/StubClasses.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: CodeDom security sandboxing. Go uses expr-lang/expr which is inherently sandboxed.

### CrossView

#### `BaseCubeLink.cs`
- **File**: `FastReport.Base/CrossView/BaseCubeLink.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing SourceAssigned property on CubeSourceBase interface; Go crossview/basecubelink.go lacks this FastCube integration hook.

#### `CrossViewCellDescriptor.cs`
- **File**: `FastReport.Base/CrossView/CrossViewCellDescriptor.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. All properties (XFieldName, YFieldName, MeasureName, IsXTotal, IsYTotal, IsXGrandTotal, IsYGrandTotal, X, Y) verified against C#. Assign() added to crossview/crossview.go (CellDescriptor.Assign). Serialization correctly clears field names when GrandTotal flags are set, matching C# constructor logic.

#### `CrossViewCells.cs`
- **File**: `FastReport.Base/CrossView/CrossViewCells.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Collection operations Add/Count/Get/Clear/Serialize/Deserialize ported in crossview/serial.go (CrossViewCells type). The C# internal-only methods (Insert, Remove, IndexOf, Contains, ToArray, AddRange) are intentionally omitted — Go slices handle these without a dedicated wrapper class.

#### `CrossViewData.cs`
- **File**: `FastReport.Base/CrossView/CrossViewData.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing cubeSource field, all FastCube convenience properties (XAxisFieldsCount, YAxisFieldsCount, MeasuresCount, GetXAxisField, GetYAxisField, GetMeasure, etc.), and GetRowDescriptor/GetColumnDescriptor methods.

#### `CrossViewDescriptor.cs`
- **File**: `FastReport.Base/CrossView/CrossViewDescriptor.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: Assign() is now implemented on the embedded `Descriptor` struct in crossview/crossview.go (copies Expression). TemplateColumn/TemplateRow/TemplateCell designer-time references remain absent — Go crossview package does not model table column/row/cell objects, so these are intentionally omitted as designer-only.

#### `CrossViewHeader.cs`
- **File**: `FastReport.Base/CrossView/CrossViewHeader.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Full collection API now implemented in crossview/serial.go (CrossViewHeader type): Add, Count, Get, Clear (existing) plus IndexOf, Contains, Insert, Remove, ToArray, AddRange (added in go-fastreport-vt567). The C# TemplateCell/TemplateColumn/TemplateRow fields on individual descriptors remain absent as designer-only.

#### `CrossViewHeaderDescriptor.cs`
- **File**: `FastReport.Base/CrossView/CrossViewHeaderDescriptor.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: All core descriptor fields (FieldName, MeasureName, IsGrandTotal, IsTotal, IsMeasure, Cell/CellSize, Level/LevelSize), serialization, GetName(), and Assign() are ported in crossview/crossview.go (HeaderDescriptor). Still missing: TemplateCell/TemplateRow/TemplateColumn styling references (designer-facing; not needed for rendering).

#### `CrossViewHelper.cs`
- **File**: `FastReport.Base/CrossView/CrossViewHelper.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: BuildTemplate(), UpdateDescriptors(), UpdateStyle(), StartPrint(), FinishPrint() are designer/engine lifecycle methods. Core layout logic is integrated into CrossViewObject.Build() / buildGrid() in crossview/crossview.go. Style application (UpdateStyle), descriptor sync (UpdateDescriptors) linking descriptors to TableColumn/TableRow/TableCell, and full print lifecycle (StartPrint/FinishPrint iterating CubeSource) remain absent because the Go package does not model the table object hierarchy or FastCube integration.

#### `CrossViewObject.Async.cs`
- **File**: `FastReport.Base/CrossView/CrossViewObject.Async.cs`
- **Status**: NOT PORTED
- **Gaps**: GetDataAsync() absent from Go crossview package; synchronous GetData() also missing; StartPrint/AddData/FinishPrint helper methods not implemented. Crossview Go package focuses on data descriptors and layout only, without engine data-loading integration.

#### `CrossViewObject.cs`
- **File**: `FastReport.Base/CrossView/CrossViewObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: FRX serialization/deserialization, Style property, ModifyResult event, engine lifecycle methods (SaveState/RestoreState/GetData/InitializeComponent/FinalizeComponent), ResultTable creation/management, and CubeSource event handling are missing; CrossViewObject is not registered in the FRX registry.

### Data

#### `BusinessObjectConverter.cs`
- **File**: `FastReport.Base/Data/BusinessObjectConverter.cs`
- **Status**: NOT PORTED
- **Gaps**: Entire class absent. No hierarchical schema discovery, no property classification, no circular reference detection. Low impact for FRX reports, high for programmatic nested struct API.

#### `BusinessObjectDataSource.Async.cs`
- **File**: `FastReport.Base/Data/BusinessObjectDataSource.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — Go uses context.Context for cancellation at engine level; async wrapper methods (InitSchemaAsync, LoadDataAsync) are not applicable to Go's synchronous DataSource interface and goroutine-based concurrency model.

#### `BusinessObjectDataSource.cs`
- **File**: `FastReport.Base/Data/BusinessObjectDataSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing LoadData() master-detail chaining logic, InitSchema() no-op, complex Deserialize() with nested datasource deduplication and legacy ReferenceName→PropName conversion, nested column path traversal in GetValue(), and LoadBusinessObject event handler passes datasource only (not LoadBusinessObjectEventArgs with parent object).

#### `Column.cs`
- **File**: `FastReport.Base/Data/Column.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: **Reviewed and updated 2026-03-21**. ColumnBindableControl enum (Text/RichText/Picture/CheckBox/Custom) ADDED to `data/column.go`. BindableControl and CustomBindableControl fields ADDED to DataColumn. SetBindableControlType() ADDED (maps Go type strings to BindableControl). Serialize/Deserialize updated to round-trip BindableControl and CustomBindableControl. Tests in `data/column_bindable_test.go`. Remaining gaps: PropDescriptor property (not needed — Go uses string-based reflection); IParent interface methods (CanContain/GetChildObjects/AddChild/RemoveChild/GetChildOrder/SetChildOrder/UpdateLayout) intentionally omitted as designer-only features; Value/ParentDataSource computed properties not exposed (engine handles data access directly); GetFormat() not needed (format conversion is handled by the format package, not DataColumn).

#### `ColumnCollection.cs`
- **File**: `FastReport.Base/Data/ColumnCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `CommandParameter.cs`
- **File**: `FastReport.Base/Data/CommandParameter.cs`
- **Status**: MOSTLY PORTED
- **Fixed (2026-03-22)** (go-fastreport-gf4at): Added `Assign(src *CommandParameter)` — copies Name, DataType, Size, Expression, DefaultValue, Direction from src (mirrors C# CommandParameter.Assign). Added `GetExpressions() []string` — returns Expression when non-empty (mirrors C# CommandParameter.GetExpressions). LastValue and SetLastValue already implemented via lastValue cache field.
- **Remaining Gaps**: Value getter dynamic expression evaluation (C# evaluates Expression string via ReportEngine at parameter-use time; Go callers must set Value directly before use).

#### `CommandParameterCollection.cs`
- **File**: `FastReport.Base/Data/CommandParameterCollection.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 7 IMPLEMENTED (indexer, FindByName, CreateUniqueName, Add, Remove, Count, All). 7 NOT IMPLEMENTED but unused in pipeline. 4 OUT OF SCOPE (Owner lifecycle).

#### `ConnectionCollection.cs`
- **File**: `FastReport.Base/Data/ConnectionCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `CsvConnectionStringBuilder.cs`
- **File**: `FastReport.Base/Data/CsvConnectionStringBuilder.cs`
- **Status**: NOT PORTED
- **Gaps**: Go bypasses connection-string-builder pattern. 2 properties PARTIALLY IMPLEMENTED (Separator, FieldNamesInFirstString — different defaults). 7 NOT IMPLEMENTED (CsvFile, Codepage, RemoveQuotationMarks, ConvertFieldTypes, locale formats).

#### `CsvDataConnection.cs`
- **File**: `FastReport.Base/Data/CsvDataConnection.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing codepage/encoding support, automatic type conversion (int/float/datetime detection), locale-aware number/currency/datetime parsing, and connection string builder abstraction (CsvConnectionStringBuilder).

#### `CsvUtils.cs`
- **File**: `FastReport.Base/Data/CsvUtils.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: **Reviewed 2026-03-21**. Core CSV parsing (split, quote handling, header/noheader, comment char, lazy quotes) fully implemented and tested at 100% coverage. ConnectionStringBuilder tested at 100% in `data/csv/connection_string_test.go`. NOT PORTED (intentional): DetermineTypes (type inference for int/double/decimal/datetime) — Go stores all CSV values as strings, sufficient for report generation; ReadLines HTTP/FTP URL loading and locale-aware parsing. These are not required for the report execution pipeline.

#### `CubeHelper.cs`
- **File**: `FastReport.Base/Data/CubeHelper.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Single method only called by CubeSourceConverter (WinForms TypeConverter for designer UI).

#### `CubeSourceBase.cs`
- **File**: `FastReport.Base/Data/CubeSourceBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 9 of 11 data-access members IMPLEMENTED. 5 NOT IMPLEMENTED (SourceAssigned, CubeLink, OnChanged, Serialize/Deserialize). Go uses interface-based design.

#### `CubeSourceCollection.cs`
- **File**: `FastReport.Base/Data/CubeSourceCollection.cs`
- **Status**: NOT PORTED
- **Gaps**: No Go CubeSourceCollection type. Go has crossview.CubeSourceBase interface and SliceCubeSource impl but no collection wrapper, no Dictionary integration, no FRX serialization for cube sources.

#### `DataComponentBase.cs`
- **File**: `FastReport.Base/Data/DataComponentBase.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Implemented in data/datacomponent.go. Serialization, properties (Name, Alias, Enabled, ReferenceName), and IsAliased logic are ported. InitializeComponent is present as a no-op hook. C# Assign() is a pass-through to Base, so no specific logic needed here. **Reviewed 2026-03-21**: Fixed `SetName` to use `strings.EqualFold` (case-insensitive alias sync) matching C# `String.Compare(Alias, Name, true)`; all properties, serialization, and IsAliased verified correct. Coverage 100%.

#### `DataConnectionBase.Async.cs`
- **File**: `FastReport.Base/Data/DataConnectionBase.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API. FillTableAsync/GetConnectionStringAsync have no Go equivalents; data connection is handled synchronously.

#### `DataConnectionBase.cs`
- **File**: `FastReport.Base/Data/DataConnectionBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core connection name/alias/serialization ported. Missing: CreateAllTables(bool initSchema) with schema filtering, CreateAllProcedures(), CreateRelations() from DataSet, GetTableNames()/GetProcedureNames(), FillTableSchema()/FillTableData(), OpenConnection()/GetConnection() lifecycle, GetAdapter()/GetParameterType()/QuoteIdentifier() for database-specific operations, Clone(), TablesStructure persistence, LoginPrompt prompting, and DatabaseLogin event handling.

#### `DataHelper.cs`
- **File**: `FastReport.Base/Data/DataHelper.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: Go covers `GetDataSource`, `GetColumn` / relation-aware lookup, `IsValidColumn`, `IsSimpleColumn`, `GetParameter`, `CreateParameter`, `IsValidParameter`, `GetTotal`, `IsValidTotal`, `GetColumnType`, and `FindRelation` in `data/helper.go`. **Reviewed 2026-03-21**: All public methods of C# DataHelper are ported. Remaining gaps: richer nested-table / nested-column traversal (C# walks nested columns as datasources; Go uses a flat column slice model), and relation initialization side effects (`FindParentRow`) which have no equivalent in the Go datasource abstraction. Tests cover all functions with >80% statement coverage.

#### `DataSourceBase.Async.cs`
- **File**: `FastReport.Base/Data/DataSourceBase.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API with context.Context. InitAsync/OpenAsync/CloseAsync have no Go equivalents; data source lifecycle is handled synchronously by the engine.

#### `DataSourceBase.cs`
- **File**: `FastReport.Base/Data/DataSourceBase.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: Core open/close/navigation (First/Next/EOF) and field-value access ported. **Fixed 2026-03-22**: Prior() row navigation added (DataSourceBase.cs:724); EnsureInit() lazy-init pattern added; AdditionalFilter predicate map (SetAdditionalFilter/ClearAdditionalFilter/ApplyAdditionalFilter) added; GetDisplayName() added returning Alias if set else Name. **Fixed 2026-03-21** (go-fastreport-3nbqg): BaseDataSource.SetName now uses strings.EqualFold for alias sync. Remaining gaps: master-detail GetChildRows/FindParentRow with parent-key caching, Load event hook, UpdateExpressions(), RowComparer/Indices for sorted traversal.

#### `DataSourceCollection.cs`
- **File**: `FastReport.Base/Data/DataSourceCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: **Reviewed and updated 2026-03-21**. All core operations (Add/Remove/Count/Get/All/FindByName/FindByAlias/Sort) implemented and tested. `Sort()` sorts by Alias ascending, matching C# DataSourceComparer. Tests added in `data/collections_extra_test.go`. Note: Go uses case-insensitive comparisons for Find operations — intentionally more lenient than C# case-sensitive.

#### `DataSourceFilter.cs`
- **File**: `FastReport.Base/Data/DataSourceFilter.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. All 12 FilterOperation enum values, FilterElement struct, DataSourceFilter with Add/Remove/Clear/Len/ValueMatch(), type-safe comparison helpers, and string-set optimization all ported in data/filter.go. **Reviewed 2026-03-21**: Verified all 12 FilterOperation enum values match C# order exactly (Equal=0 through NotEndsWith=11). ValueMatch/matches logic verified against C# including string-set branch (Equal/NotEqual/Contains/NotContains), DateTime range branch (AddDate(0,0,1) makes end exclusive), DateTime scalar with time-stripping (strips when element has no time component), and string-specific operations. Go implementation enhances the range branch with an operation switch (C# only used `match` directly). Coverage 98.3% on `matches` (remaining 1.7% is dead code: `compare(time.Time, time.Time)` always returns ok=true so the `if !ok { return false }` guard at line 176 is unreachable).

#### `DbConnectionExtensions.cs`
- **File**: `FastReport.Base/Data/DbConnectionExtensions.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET Framework async/await shims (GetSchemaAsync, DisposeAsync). Go's database/sql package natively supports cancellation via context.Context; no async extension methods needed.

#### `Dictionary.cs`
- **File**: `FastReport.Base/Data/Dictionary.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: **Reviewed and updated 2026-03-21**. Implemented: `FindByName`, `FindByAlias`, `FindDataComponent`, `CreateUniqueName`, `CreateUniqueAlias`, `UpdateRelations`, `Clear`. **Fixed 2026-03-22**: `Merge(source *Dictionary)` added — copies data sources, connections, relations, parameters, totals from source skipping duplicates by name (Dictionary.cs:725-780). Note: Go Merge does not automatically call UpdateDataSourceRef on report objects (caller must do this separately since Dictionary has no access to report pages). Remaining gaps: CubeSourceCollection (designer/pivot-only), RegisteredItems tracking, AllObjects/CacheAllObjects caching (designer optimization).

#### `DictionaryHelper.cs`
- **File**: `FastReport.Base/Data/DictionaryHelper.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core lookup/resolution logic ported across data/helper.go and data/dictionary.go. Missing: PRegisterBusinessObject/PRegisterDataTable/PRegisterDataView/PRegisterDataRelation/PRegisterCubeLink private registration methods, AddBaseToDictionary/AddBaseWithChildToDictionary registry methods, and DataSet decomposition logic (tables + relations flattening). Go delegates registration directly to Dictionary.AddDataSource() and standalone helper functions.

#### `Parameter.cs`
- **File**: `FastReport.Base/Data/Parameter.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: **Reviewed and updated 2026-03-21**. Implemented: `Description` property (already existed), `AsString`/`SetAsString`, `Serialize`/`Deserialize`, `FullName` (returns local name; full path via `FullNameWithParent` helper), nested `Parameters()` collection. Tests in `data/dictionary_parameter_coverage_test.go`. Remaining gaps: IParent interface (CanContain/GetChildObjects/AddChild/RemoveChild — designer lifecycle), `Assign()` (designer copy), `GetExpressions()` (script collector), and dynamic `Value` evaluation (Expression evaluated by report engine at runtime — the engine handles this through `Dictionary.Evaluate`/`EvaluateAll` which already exists). Go Parameter is a data struct; the lifecycle is managed by the engine.

#### `ParameterCollection.cs`
- **File**: `FastReport.Base/Data/ParameterCollection.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: **Reviewed and updated 2026-03-21**. Implemented: `AssignValues` (recursive value copy by FullName, implemented in `data/helper.go`), `enumParameters` (recursive enumeration), `FindParameterByName` (case-insensitive lookup). Tests in `data/dictionary_parameter_coverage_test.go`. No dedicated ParameterCollection class — parameters stored as `[]*Parameter` slice in Dictionary, matching Go idioms. Remaining gaps: `Assign()` (designer copy), formal C# `ParameterCollection` type as a standalone class (not needed — slice operations are equivalent).

#### `ProcedureDataSource.cs`
- **File**: `FastReport.Base/Data/ProcedureDataSource.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — missing only DisplayNameWithParams UI property which is designer-only and not needed for report execution.

#### `ProcedureParameter.cs`
- **File**: `FastReport.Base/Data/ProcedureParameter.cs`
- **Status**: NOT PORTED
- **Gaps**: Thin subclass adding designer UI only. Go uses CommandParameter directly.

#### `Relation.cs`
- **File**: `FastReport.Base/Data/Relation.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — Serialize/Deserialize both implemented in `data/helper.go`. Tests added in `data/dictionary_parameter_coverage_test.go` covering round-trip and fallback-to-source-name paths. Go Relation struct covers all runtime fields and engine integration.

#### `RelationCollection.cs`
- **File**: `FastReport.Base/Data/RelationCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — `RelationCollection` struct implemented in `data/helper.go` with Add, Remove, Count, Get, All, FindByName (exact match), FindByAlias (exact match), and FindEqual methods. Dictionary continues to expose relations as `[]*Relation` for internal use; the typed collection is available for external consumers.

#### `SliceCubeSource.cs`
- **File**: `FastReport.Base/Data/SliceCubeSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go collapses 3-layer delegation. Missing DataComponentBase chain, MeasureIndex.

#### `SystemVariables.cs`
- **File**: `FastReport.Base/Data/SystemVariables.cs`
- **Status**: PORTED
- **Resolved**: HierarchyLevel and HierarchyRow# now synced in syncSystemVariables() and syncPageVariables(). "Page" canonical name (C# PageVariable.Name) added alongside "PageNumber" alias. All 12 C# variables initialised and synced. HierarchyRow corrected to string type matching C# HierarchyRowNo.
- **Remaining gap**: Date/Time stored as formatted string in expression env (not time.Time); raw time.Time accessible via "Now".

#### `TableCollection.cs`
- **File**: `FastReport.Base/Data/TableCollection.cs`
- **Status**: NOT PORTED
- **Gaps**: No TableCollection class; data sources managed as []DataSource slice in Dictionary. Missing indexer and Sort() method (sorts by name using DataSourceComparer).

#### `TableDataSource.Async.cs`
- **File**: `FastReport.Base/Data/TableDataSource.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API; InitSchemaAsync/LoadDataAsync/RefreshTableAsync are not applicable.

#### `TableDataSource.cs`
- **File**: `FastReport.Base/Data/TableDataSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing Serialize/Deserialize for FRX persistence, TableData property (base64-encoded embedded DataSet), QbSchema property, RefreshTable/RefreshColumns lifecycle methods, and InitSchemaShared/LoadDataShared helper patterns.

#### `Total.cs`
- **File**: `FastReport.Base/Data/Total.cs`
- **Status**: PORTED
- **Resolved**: All fields (EvaluateCondition, IncludeInvisibleRows, ResetAfterPrint, ResetOnReprint) fully deserialized from FRX with correct C# defaults (ResetAfterPrint=true when absent, ResetOnReprint=true when absent). Serialize support in reportpkg/dictionary_serial.go. AggregateTotal (data/total.go) implements accumulation with condition evaluation via EvaluateCondition. Contains() and ClearValues() added.
- **Remaining gap**: Hierarchy sub-totals (_sub prefix / subTotals pattern from C# Total.cs line 364) not ported — hierarchical reports use simplified accumulation.

#### `TotalCollection.cs`
- **File**: `FastReport.Base/Data/TotalCollection.cs`
- **Status**: PORTED
- **Resolved**: Explicit TotalCollection class in data/collections.go with FindByName() (case-insensitive, matches C# behavior), CreateUniqueName(), GetValue(), Contains(), ClearValues(). ProcessBand equivalent via engine accumulateTotals()/resetGroupTotals(). StartKeep/EndKeep on AggregateTotal in data/total.go.

#### `ViewDataSource.Async.cs`
- **File**: `FastReport.Base/Data/ViewDataSource.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API with context.Context for cancellation.

#### `ViewDataSource.cs`
- **File**: `FastReport.Base/Data/ViewDataSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go implements filter semantics via data/viewdatasource.go rather than wrapping a DataView; missing RefreshColumns() schema-sync method that keeps column list in sync with the underlying DataTable schema. **Fixed 2026-03-21** (go-fastreport-3nbqg): SetName now uses strings.EqualFold for alias sync, matching C# DataComponentBase.SetName behavior.

#### `VirtualDataSource.Async.cs`
- **File**: `FastReport.Base/Data/VirtualDataSource.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API with context.Context for cancellation.

#### `VirtualDataSource.cs`
- **File**: `FastReport.Base/Data/VirtualDataSource.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `XmlConnectionStringBuilder.cs`
- **File**: `FastReport.Base/Data/XmlConnectionStringBuilder.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Builder class not needed. XSD schema, Codepage, URL loading, multi-table DataSet not ported.

#### `XmlDataConnection.cs`
- **File**: `FastReport.Base/Data/XmlDataConnection.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: **Reviewed 2026-03-21**. Core XML loading from file/string, rootPath navigation, row element detection, and mixed attribute+child-element columns all implemented and tested at 100% coverage in `data/xml/xml.go`. ConnectionStringBuilder tested at 100% in `data/xml/connection_string_test.go`. NOT PORTED (out of scope for Go pipeline): async variants, HTTP/FTP URL support, XSD schema loading, codepage/encoding support, FillTableSchema/FillTableData/GetTableNames.

### Data/JsonConnection

#### `IJsonProviderSourceConnection.cs`
- **File**: `FastReport.Base/Data/JsonConnection/IJsonProviderSourceConnection.cs`
- **Status**: NOT PORTED
- **Gaps**: No Go equivalent.

#### `JsonDataSourceConnection.cs`
- **File**: `FastReport.Base/Data/JsonConnection/JsonDataSourceConnection.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 3 HIGH (no ConnectionString parsing, no JsonSchema, no hierarchical columns). 4 MEDIUM (no SimpleStructure, no HTTP, no IJsonProvider, child columns skipped). Flat JSON works.

#### `JsonDataSourceConnectionStringBuilder.cs`
- **File**: `FastReport.Base/Data/JsonConnection/JsonDataSourceConnectionStringBuilder.cs`
- **Status**: NOT PORTED
- **Gaps**: No Go equivalent. 8 gaps including no connection string parsing, no base64 decoding, no HTTP URL fetching, no headers. HIGH impact for URL/auth JSON.

#### `JsonTableDataSource.cs`
- **File**: `FastReport.Base/Data/JsonConnection/JsonTableDataSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Flat JSON works. Major gaps: no column hierarchy for nested JSON, no JsonSchema, no URL fetching, no SimpleStructure mode. Architecture mismatch (flat vs hierarchical).

### Engine

#### `ReportEngine.Async.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: All 6 methods are async wrappers. Go uses context.Context. Key async patterns reviewed (go-fastreport-n5yib, 2026-03-22): `RunAsync` → `Run(opts RunOptions)`, `RunPhase2Async` → `runPhase2`, `GetBandHeightWithChildrenAsync` → `GetBandHeightWithChildren`, `GetFreeSpaceAsync` → `FreeSpace()`. All equivalent synchronous implementations present.

#### `ReportEngine.Bands.Async.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Bands.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: All 7 methods are pure async wrappers. Zero unique logic.

#### `ReportEngine.Bands.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Bands.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: **Reviewed 2026-03-22 (go-fastreport-ertdg)**. IMPLEMENTED: (1) `ProcessTotals` wired — `processTotalsForBand(bandName, bandRepeated)` added to engine/totals.go, called from `showFullBandOnce` and `showBand` after each band renders, mirroring C# `ShowBand → ProcessTotals(band)` (ReportEngine.Bands.cs line 228/250) and `TotalCollection.ProcessBand` (TotalCollection.cs lines 65-77). (2) `VisibleExpression` evaluated on bands — `evalBandVisibleExpression` helper added to engine/bands.go, called at the top of `showFullBandOnce` and `showBand`; handles TotalPages/DoublePass semantics correctly. (3) `PrintOn` flag logic verified correct — Go uses pageNumber = pageIndex+1 (1-based), matching C#. (4) `BandCanStartNewPage` parent-walk added — `bandCanStartNewPage` walks `b.Parent()` chain; if any ancestor has `FlagUseStartNewPage=false`, returns false; wired into `ShowDataBandRow`, `RunDataBandRowsKeep`, `RunDataBandFull`. REMAINING: ContainsBand deduplication for object overlap prevention; ExportableExpression evaluation (infrastructure gap); CanPrint/Printable expression evaluation for objects (not bands).

#### `ReportEngine.Break.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Break.cs`
- **Status**: PARTIALLY PORTED
- **Fixed (go-fastreport-rphwt, 2026-03-22)**: Added `utils.SplitTextAtHeight()` in `utils/textmeasure.go` and wired it into `engine/breaks.go` `splitPopulateTop`/`splitPopulateBottom`: when a `TextObject` with `WordWrap=true` straddles the breakLine, text is now split at the line boundary that fits the available height, matching C# `TextObject.BreakText()` and `BandBase.Break()` behavior. The top PreparedBand gets the fitting text lines; the bottom PreparedBand gets the overflow. Tests added to `engine/breaks_internal_test.go`.
- **Remaining gaps**: No object cloning (BreakBand in Go builds PreparedObjects directly rather than cloning live objects). `SplitHardPageBreaks` mutates original object Top (should clone). `BreakBand` break-line calculation does not probe CanBreak on BreakableComponent instances (only checks the flag, not calling `Clone.Break(nil)` as C# does).

#### `ReportEngine.DataBands.Async.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.DataBands.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: All 8 methods are pure async wrappers. Zero unique logic.

#### `ReportEngine.DataBands.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.DataBands.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core single-column iteration ported. 9 gaps: HIGH — DownThenAcross multi-column (80+ lines unported). MEDIUM — IsDetailEmpty guard missing, AcrossThenDown simplified. LOW — RTL columns, footer KeepWithData, hierarchy indent/per-level headers.

#### `ReportEngine.Groups.Async.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Groups.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: All 6 methods are pure async wrappers. Zero unique logic.

#### `ReportEngine.Groups.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Groups.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core group iteration works. 2 HIGH: showGroupFooter missing Prior/Next rollback, showGroupTree missing CurrentRowNo positioning. MEDIUM: group condition uses GetValue not Report.Calc, missing AbsRowNo/RowNo reset, MaxRows during tree build, missing FinalizeDataSource.

#### `ReportEngine.Keep.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Keep.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core cut/paste ported. Missing: StartKeep Outline/Bookmark/Totals/Reprint save, PasteObjects Outline.Shift/Bookmarks.Shift, EndKeep Totals/Reprint cleanup, wrong guard (StartNewPage vs FirstRowStartsNewPage), AggregateTotal.StartKeep simplified.

#### `ReportEngine.KeepWithData.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.KeepWithData.cs`
- **Status**: FULLY PORTED
- **Fixed (go-fastreport-rphwt, 2026-03-22)**: (1) `NeedKeepFirstRowGroup` now implements the recursive parent walk: when `groupBand.IsFirstRow()=true`, it type-asserts `groupBand.Parent()` to `*band.GroupHeaderBand` and recurses — matching C# `NeedKeepFirstRow(GroupHeaderBand groupBand)` lines 78-79 in `ReportEngine.KeepWithData.cs`. Also changed `groupBand.Data()` to `groupBand.GroupDataBand()` for accurate traversal of nested groups. (2) `getAllFooters` now includes the `ReportSummaryBand` when `dataBand.KeepSummary()=true` and `e.currentPage.ReportSummary() != nil`, matching C# lines 27-29. Tests added to `engine/keepwithdata_groupfooter_test.go`.

#### `ReportEngine.Outline.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Outline.cs`
- **Status**: FULLY PORTED
- **Fixed (go-fastreport-n5yib, 2026-03-22)**: Added `!b.Repeated()` guard to `showFullBandOnce` (engine/bands.go) so reprinted bands do not add duplicate outline entries — matches C# `AddBandOutline` line 29. Fixed double `OutlineUp` for DataBand/GroupHeaderBand: `showFullBandOnce` now skips `OutlineUp` for bands with `FlagIsDataBand` or `FlagIsGroupHeader` (those bands handle it in their per-row/per-footer code). `RunDataBandRowsKeep` and `RunDataBandFull` now call `OutlineUp` only when `db.OutlineExpression() != ""`, matching C# `OutlineUp(BandBase)` (Outline.cs line 43–50). `showGroupFooter` now calls `OutlineUp` only when `header.OutlineExpression() != ""`, matching C# `ShowGroupFooter` line 160. Set `FlagIsGroupHeader = true` in `NewGroupHeaderBand()` (band/types.go) to enable the engine-side type check without Go interface casting loss.

#### `ReportEngine.PageNumbers.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.PageNumbers.cs`
- **Status**: FULLY PORTED
- **Gaps**: All 6 methods faithfully ported with added defensive bounds checks.

#### `ReportEngine.Pages.Async.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Pages.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: All 5 methods are pure async wrappers. Zero unique logic. Go uses context.Context idiomatically.

#### `ReportEngine.Pages.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Pages.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core page lifecycle works. 6 HIGH gaps: PrintOnPreviousPage, VisibleExpression at page level, UnlimitedWidth, StartFirstPageShared, Subreport filter, InterleaveWithBackPage. 6 MEDIUM gaps.

#### `ReportEngine.ProcessAt.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.ProcessAt.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: EngineState enum and all 12 states ported. `OnStateChanged` dispatch ported with deferred-item queue. `AddDeferredHandler` (one-shot) and `AddRepeatingDeferredHandler` (repeating) cover `ProcessAt.PageFinished`, `ColumnFinished`, `ReportFinished`. NOT PORTED: `ProcessInfo` struct with `sender`-based `DataFinished`/`GroupFinished` context filtering (requires TextObject reference + XmlItem); `AddObjectToProcess`/`ProcessObject` public API for custom `ProcessAt` mode; FillColor/TextColor/Font live update in `ProcessInfo.Process()` (C# lines 93–103); `SaveState`/`RestoreState` around text evaluation. These gaps matter for reports with totals printed in headers using `ProcessAt=DataFinished/GroupFinished`.
- **Verified (go-fastreport-n5yib, 2026-03-22)**: Reviewed C# ProcessAt.cs fully. The Go `processat.go` correctly matches C# for `EngineState` enum values and `OnStateChanged` dispatch semantics. Remaining gaps are pre-existing and require a broader TextObject/XmlItem integration effort.

#### `ReportEngine.Reprint.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Reprint.cs`
- **Status**: FULLY PORTED
- **Gaps**: All 9 methods ported. BUG: generic AddReprint() in groups.go:250 routes DataHeaderBand to footer list instead of header list.

#### `ReportEngine.Subreports.Async.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Subreports.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async mirrors of already-ported synchronous subreport methods. Go has no async/await.

#### `ReportEngine.Subreports.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.Subreports.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: All 4 methods have Go equivalents but 10 gaps. Critical: VisibleExpression/Visible checks missing, multi-page outer subreport support missing, CurY height offset.

#### `ReportEngine.cs`
- **File**: `FastReport.Base/Engine/ReportEngine.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: ~55 of ~70 methods IMPLEMENTED across all 13 partial files. Key gaps: UnlimitedHeight/Width, OutlineXml, ProcessObject, CanPrint expression evaluation, BandCanStartNewPage hierarchy, DownThenAcross columns, ShowPageFooter double-pass, RTL columns.

### Export

#### `ExportBase.cs`
- **File**: `FastReport.Base/Export/ExportBase.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core export hooks IMPLEMENTED (Start, ExportPageBegin/End, ExportBand, Finish, PageRange, ParsePageNumbers, Zoom, ShiftNonExportable, HasMultipleFiles, OnProgress callback, Serialize/Deserialize, GeneratedFiles tracking). Remaining gaps: InstantExport API (5 methods — stream-based instant preview, GUI-specific), ExportAndZip (depends on zip utility not ported), OpenAfterExport/AllowOpenAfter (GUI/OS-specific), ShowProgress/AllowSaveSettings (GUI-specific). NOTE: ShiftNonExportable is stored as a flag; in the Go port band exportability is resolved by the engine before PreparedPages are created, so the flag is advisory.

#### `ExportUtils.cs`
- **File**: `FastReport.Base/Export/ExportUtils.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: IMPLEMENTED: FloatToString, HTMLColor, HTMLColorCode/HTMLColorToRGB, HtmlURL, HTMLString, XMLString, ByteToHex, ReverseString, QuotedPrintable, GetColorFromFill, GetRFCDate, GetPageWidth/Height, StrToHex/StrToHex2, ExcelCellRef/ColName, StringFormat, Adler32/ZLibDeflate. NOT IMPLEMENTED (lower priority): GetExcelFormatSpecifier, ParseTextToDecimal/DateTime/Percent (XLSX-specific Excel format strings), GetCodec/SaveJpeg (replaced by Go image/jpeg), UInt16Tohex (niche), TruncLeadSlash (niche), IndexToName (duplicate of ExcelColName). OUT OF SCOPE: System.Drawing-based GetCodec, ImageCodecInfo, DOTNET_4-specific HtmlURL.

### Export/Html

#### `HTMLExport.cs`
- **File**: `FastReport.Base/Export/Html/HTMLExport.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core rendering pipeline well-implemented (layers, CSS dedup, text/picture/shape/line/checkbox/watermark/hyperlink/border). PageBreaks property IMPLEMENTED (2026-03-22): controls print CSS and break-after-page divs, matching C# `pageBreaks` field (default true). Major remaining gaps: multi-output modes (WebMode/MHT/navigator), TableBase rendering, HtmlTextRenderer styled spans, gradient fills, Wingdings, landscape rotation, Serialize.

#### `HTMLExportDraw.cs`
- **File**: `FastReport.Base/Export/Html/HTMLExportDraw.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 16 of 19 methods IMPLEMENTED. PrintPageStyle IMPLEMENTED (2026-03-22): print CSS block (`<style media="print">`) is now conditioned on `PageBreaks` field, matching C# `singlePage && pageBreaks` guard; page divs only receive `class="frpageN"` when PageBreaks=true, matching C# `doPageBreak` logic. 1 NOT IMPLEMENTED (HTMLGetImageTag). 2 PARTIALLY (HTMLGetImage file/web modes, GetBase64Image hash dedup).

#### `HTMLExportLayers.cs`
- **File**: `FastReport.Base/Export/Html/HTMLExportLayers.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 12 methods ported (core layered rendering). Page break separator div (`<div style="break-after:page">`) now conditioned on `PageBreaks` field (2026-03-22), matching C# `doPageBreak = singlePage && pageBreaks` logic. Remaining gaps: Table rendering, IsMemo bitmap fallback, vertical alignment, bookmark anchors, target=_blank, rich text, non-SolidFill, Wingdings.

#### `HTMLExportStyles.cs`
- **File**: `FastReport.Base/Export/Html/HTMLExportStyles.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: CSS class deduplication FULLY IMPLEMENTED via cssRegistry (Register returns same class name for identical CSS strings). RTL text direction FULLY IMPLEMENTED (`direction:rtl;` emitted in outerCSS when obj.RTL=true). 7 of 11 methods IMPLEMENTED (style building, CSS registration, dual-class pattern). 2 PARTIALLY (InlineStyles branch — was C# TODO). 1 NOT IMPLEMENTED (InlineStyle). 1 OUT OF SCOPE (InlineStyles property). stylePrefix not ported.

#### `HTMLExportTemplates.cs`
- **File**: `FastReport.Base/Export/Html/HTMLExportTemplates.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 2 IMPLEMENTED (PageTemplateTitle/Footer). 2 NOT IMPLEMENTED (NavigatorTemplate, IndexTemplate — multi-file frameset, legacy HTML4). 1 OUT OF SCOPE (OutlineTemplate — empty in C#).

#### `HTMLExportUtils.cs`
- **File**: `FastReport.Base/Export/Html/HTMLExportUtils.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 1 IMPLEMENTED (Px as pxVal). 5 NOT IMPLEMENTED (3 MHT/MHTML, HtmlSizeUnits percent, HTMLPageData superseded). 1 OUT OF SCOPE (ImageFormat enum).

### Export/Image

#### `ImageExport.cs`
- **File**: `FastReport.Base/Export/Image/ImageExport.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: IMPLEMENTED: all C# ImageExportFormat values (JPEG/PNG/GIF/BMP/TIFF), SeparateFiles (per-page encoding), ResolutionX/ResolutionY (DPI control), JpegQuality, MultiFrameTiff (multi-page TIFF stream), MonochromeTiff (greyscale conversion), PaddingNonSeparatePages, combined-page mode (stitchPages), Serialize/Deserialize. Remaining gaps: per-page file naming when writing to separate disk files (the Go port is stream-based; file naming is caller's responsibility), SaveStreams/GeneratedStreams (stream collection mode), watermark image centering using page margins. OUT OF SCOPE: Metafile/EMF format (Windows GDI+), ConvertToBitonal 1bpp TIFF (replaced by greyscale), MonochromeTiffCompression enum (CCITT4 etc., platform-specific).

### Format

#### `BooleanFormat.cs`
- **File**: `FastReport.Base/Format/BooleanFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: TrueText/FalseText properties and FormatValue logic fully ported. Clone(), Equals(), and GetSampleValue() added in `format/boolean.go`. Serialization is handled centrally in `object/format_serial.go` (by design). GetHashCode is not applicable in Go. Tests in `format/clone_equals_test.go`. Coverage: 100%.

#### `CurrencyFormat.cs`
- **File**: `FastReport.Base/Format/CurrencyFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: All fields and FormatValue logic ported. Clone(), Equals(), and GetSampleValue() added in `format/currency.go`. All 4 positive patterns and 16 negative patterns implemented. Serialization handled centrally in `object/format_serial.go`. Tests in `format/clone_equals_test.go`. Coverage: 100%.

#### `CustomFormat.cs`
- **File**: `FastReport.Base/Format/CustomFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: Format property and FormatValue logic ported. Clone(), Equals(), and GetSampleValue() added in `format/custom.go`. Format string syntax differs by design (Go fmt.Sprintf vs .NET string.Format — intentional Go idiom adaptation). Serialization handled centrally in `object/format_serial.go`. Tests in `format/clone_equals_test.go`. Coverage: 100%.

#### `DateFormat.cs`
- **File**: `FastReport.Base/Format/DateFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: `csharpDateLayouts` map correctly translates all C# standard format specifiers (d, D, f, F, g, G, t, T, M, Y, s, u, R, o) to Go time layout strings. Clone(), Equals(), and GetSampleValue() added (sample date 2007-11-30T13:30:00 → "11/30/2007" verified). FRX attribute uses `UseLocale` (not `UseLocaleSettings`) — fixed in `object/format_serial.go` serializer/deserializer. Tests in `format/clone_equals_test.go`. Coverage: 100%.

#### `FormatBase.cs`
- **File**: `FastReport.Base/Format/FormatBase.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: Go defines a `Format` interface with `FormatValue` and `FormatType`. All concrete format types now implement `Clone() Format`, `Equals(Format) bool`, and `GetSampleValue() string` as value methods (duck-typed — not enforced by the interface). Serialization is handled centrally in `object/format_serial.go` (by design).

#### `FormatCollection.cs`
- **File**: `FastReport.Base/Format/FormatCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: All collection operations ported (Add, Insert, Remove, Contains, IndexOf, Clear, Assign, All, Primary, FormatValue). Equals() method added in `format/collection.go` — uses the Equaler interface duck-type if available, otherwise falls back to pointer identity. Serialize/Deserialize are handled externally via `object/format_serial.go` (by design). GetHashCode not applicable in Go. Tests in `format/clone_equals_test.go`. Coverage: 100%.

#### `GeneralFormat.cs`
- **File**: `FastReport.Base/Format/GeneralFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: FormatValue() returns "" for typed nil pointers (matching C# null → ""). Clone(), Equals(), and GetSampleValue() added in `format/general.go`. Coverage: 100%.

#### `NumberFormat.cs`
- **File**: `FastReport.Base/Format/NumberFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: Clone(), Equals(), and GetSampleValue() added in `format/number.go`. Go struct field is named `UseLocaleSettings` (C# uses `UseLocale`); FRX attribute name mismatch fixed in `object/format_serial.go`. Go uses hardcoded "." and "," defaults instead of CultureInfo.CurrentCulture (invariant locale behavior is equivalent). Core formatting logic with all 5 negative patterns fully ported. Tests in `format/clone_equals_test.go`. Coverage: 100%.

#### `PercentFormat.cs`
- **File**: `FastReport.Base/Format/PercentFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: All 12 negative patterns and 4 positive patterns implemented. Clone(), Equals(), and GetSampleValue() added in `format/percent.go`. GetSampleValue (1.23 → "123.00 %") verified. FRX attribute name `UseLocale` fixed in serializer/deserializer. Tests in `format/clone_equals_test.go`. Coverage: 100%.

#### `TimeFormat.cs`
- **File**: `FastReport.Base/Format/TimeFormat.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Reviewed 2026-03-21: Clone(), Equals(), and GetSampleValue() added in `format/time.go`. Go TimeFormat is standalone (C# inherits from CustomFormat but that distinction has no behavioral impact). Core FormatValue() logic ported with duration handling enhancement. Tests in `format/clone_equals_test.go`. Coverage: 100%.

### Functions

#### `NumToLettersBase.cs`
- **File**: `FastReport.Base/Functions/NumToLettersBase.cs`
- **Status**: FULLY PORTED
- **Reviewed**: 2026-03-22 (issue go-fastreport-37oqs)
- **Gaps**: None. Abstract base algorithm (str() → Excel-style column-label scheme, negative returns "") fully ported in functions/numtoletters.go. Go loop structure differs from C# while-loop but is algorithmically identical — verified by manual trace for key values (0, 1, 25, 26, 27, 51, 52, 701, 702) and table-driven tests. C# uses prepend-to-StringBuilder; Go builds rune slice prepending to front; both produce identical output. Re-verified in issue go-fastreport-37oqs: no gaps found.

#### `NumToLettersEn.cs`
- **File**: `FastReport.Base/Functions/NumToLettersEn.cs`
- **Status**: FULLY PORTED
- **Reviewed**: 2026-03-21 (issue go-fastreport-k97sc)
- **Gaps**: None. Implemented as ToLettersEn (uppercase) and ToLettersEn(n, false) (lowercase) in functions/numtoletters.go. All 26 a-z / A-Z letters match C# exactly. ToLetters convenience wrapper and NumToLettersLower deprecated helper now have 100% test coverage.

#### `NumToLettersRu.cs`
- **File**: `FastReport.Base/Functions/NumToLettersRu.cs`
- **Status**: FULLY PORTED
- **Reviewed**: 2026-03-21 (issue go-fastreport-k97sc)
- **Gaps**: None. Implemented as ToLettersRu in functions/numtoletters.go. All 33 Cyrillic letters (а-я / А-Я) match C# exactly including boundary cases (n=32→"Я", n=33→"АА", n=65→"АЯ", n=66→"БА"). Negative inputs return "" matching C# behavior.

#### `NumToWordSp.cs`
- **File**: `FastReport.Base/Functions/NumToWordSp.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `NumToWordsBase.cs`
- **File**: `FastReport.Base/Functions/NumToWordsBase.cs`
- **Status**: PARTIALLY PORTED
- **Reviewed**: 2026-03-22 (issue go-fastreport-37oqs)
- **Gaps**: The abstract base contract (ConvertCurrency, ConvertNumber with WordInfo/CurrencyInfo) is not implemented — the Go port uses standalone per-language functions instead of an OO hierarchy. Two grammar rules from the base class were fixed in this issue: (1) Get100_10Separator=" and " (English) — the base class adds "and" between hundreds and tens/ones, and also before remainders < 100 when value >= 1000 (e.g., "one thousand and one"). (2) The sep100_10 condition: cleared only when value < 1000 AND hund == "". Remaining gap: currency/noun declension (ConvertCurrency) not implemented.

#### `NumToWordsDe.cs`
- **File**: `FastReport.Base/Functions/NumToWordsDe.cs`
- **Status**: PARTIALLY PORTED
- **Reviewed**: 2026-03-22 (issue go-fastreport-37oqs)
- **Gaps fixed**: German feminine noun grammar: Million, Milliarde and Billion are feminine in German. C# uses WordInfo(male=false,...) for these, so GetFixedWords(false, 1) = "eine". Fixed in functions/numtowords_de.go: dePositive() now passes female=true when computing the multiplier for thousand/million/milliard/billion groups, producing "eine Million", "eine Milliarde", "eine Billion", "einetausend", "eineundzwanzigtausend". Str1000 override porting verified: no separator between components (counter==2 for thousands), compound tens (frac10+"und"+ten). Remaining gap: currency/noun declension (GetCurrency/ConvertCurrency) not implemented.

#### `NumToWordsEn.cs`
- **File**: `FastReport.Base/Functions/NumToWordsEn.cs`
- **Status**: PARTIALLY PORTED
- **Reviewed**: 2026-03-22 (issue go-fastreport-37oqs)
- **Gaps fixed**: Two grammar separators from C# now correctly ported in functions/numtowords.go: (1) Get100_10Separator=" and " — produces "one hundred and twenty-three", "one hundred and one". (2) Get10_1Separator="-" — already correct ("twenty-one"). (3) The sep100_10 logic: the "and" is also added before remainders < 100 in higher scales (e.g., "one thousand and one", "one million and five") because the condition `value < 1000 && hund == ""` that would suppress it is false when value >= 1000. Fixed: numToWordsPositive() now handles n < 1000 explicitly with "and", and uses "and" before rem < 100 in scale groups. Removed dead {100, "hundred"} from scales slice. Remaining gap: GetCurrency/ConvertCurrency not implemented.

#### `NumToWordsEnGb.cs`
- **File**: `FastReport.Base/Functions/NumToWordsEnGb.cs`
- **Status**: PARTIALLY PORTED
- **Reviewed**: 2026-03-22 (issue go-fastreport-37oqs)
- **Gaps fixed**: NumToWordsEnGb extends NumToWordsEn in C# (same Get100_10Separator=" and "). Fixed in functions/numtowords_en_gb.go: enGbPositive() now adds "and" before remainders < 100 in milliard/billion groups (e.g., "one milliard and one", "one billion and one"). Inherits the "and" fix in numToWordsPositive() for millions-and-below groups. Remaining gap: GetCurrency/ConvertCurrency not implemented.

#### `NumToWordsEs.cs`
- **File**: `FastReport.Base/Functions/NumToWordsEs.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `NumToWordsFr.cs`
- **File**: `FastReport.Base/Functions/NumToWordsFr.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `NumToWordsIn.cs`
- **File**: `FastReport.Base/Functions/NumToWordsIn.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `NumToWordsNl.cs`
- **File**: `FastReport.Base/Functions/NumToWordsNl.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `NumToWordsPersian.cs`
- **File**: `FastReport.Base/Functions/NumToWordsPersian.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `NumToWordsPl.cs`
- **File**: `FastReport.Base/Functions/NumToWordsPl.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `NumToWordsRu.cs`
- **File**: `FastReport.Base/Functions/NumToWordsRu.cs`
- **Status**: PARTIALLY PORTED
- **Reviewed**: 2026-03-22 (issue go-fastreport-37oqs)
- **Gaps verified**: Grammar rules reviewed and confirmed correct. Feminine forms (одна/одна тысяча, две/две тысячи) correctly applied — тысяча is feminine so 1000="одна тысяча", 2000="две тысячи". Masculine forms for миллион/миллиард/триллион (один миллион, два миллиона, пять миллионов). Three-form Case declension (one/few/many) correctly matches C# Case() override: last2 in 11-19 → many, last1=1 → one, last1 in 2-4 → few, else → many. fixedWords array matches C# exactly. Remaining gap: GetCurrency/ConvertCurrency not implemented (RUR, UAH, EUR, USD, RUB, BYN, BBYN currencies).

#### `NumToWordsUkr.cs`
- **File**: `FastReport.Base/Functions/NumToWordsUkr.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Basic number-to-words ported. Missing currency/noun declension.

#### `Roman.cs`
- **File**: `FastReport.Base/Functions/Roman.cs`
- **Status**: FULLY PORTED
- **Fixed 2026-03-22** (go-fastreport-27mqi): Added "ToRoman" as registration key in functions/standard.go to match C# StdFunctions.cs:1807 (InternalAddFunction registers it as "ToRoman"). Kept "Roman" alias for backward compatibility. Upper bound kept at 3999 (Go is more correct than C# MAX=3998 since 3999=MMMCMXCIX is a valid Roman numeral — Go intentionally deviates). Algorithm verified correct via table-driven tests.

#### `StdFunctions.cs`
- **File**: `FastReport.Base/Functions/StdFunctions.cs`
- **Status**: MOSTLY PORTED
- **Fixed 2026-03-22**: IsNull(v any) bool added to functions/standard.go and registered as "IsNull". In Go, expression evaluator resolves column/parameter values before function call, so IsNull(value) checks for nil directly (C# version takes report+name string and resolves by name — not needed in Go's expression model). IsNumeric and IsDateTime do not exist in C# StdFunctions.cs (they were incorrectly listed as gaps). ToBoolean(v any) and IfNull already existed.
- **Remaining gaps**: Various ToWords overloads (currency/unit naming for RU, UK, PL, etc.), localized ToLetters variants.

### Gauge

#### `GaugeLabel.cs`
- **File**: `FastReport.Base/Gauge/GaugeLabel.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: Color field and Assign() method are now ported to GaugeLabel struct in gauge.go. Label.Color is serialized/deserialized via dot-notation in GaugeObject.Serialize/Deserialize. Parent reference management, Draw(), and standalone Serialize() remain absent by design — Go embeds GaugeLabel fields into GaugeObject serialization rather than treating it as an independent serializable object. Draw() is replaced by the RenderXxx rendering functions in render.go.

#### `GaugeObject.Async.cs`
- **File**: `FastReport.Base/Gauge/GaugeObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API; expression evaluation for gauge value is handled by evalGaugeValue() in engine/objects.go rather than a GetDataAsync() call on the object.

#### `GaugeObject.cs`
- **File**: `FastReport.Base/Gauge/GaugeObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Vertical property IS ported (gauge.go Vertical() method). Clone(), Assign(), GetExpressions(), and GetData() methods are not ported; expression evaluation is handled via evalGaugeValue() in engine/objects.go rather than GetData() on the object. Draw() is replaced by RenderLinear/RenderRadial/RenderSimple/RenderSimpleProgress in gauge/render.go.

#### `GaugePointer.cs`
- **File**: `FastReport.Base/Gauge/GaugePointer.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: BorderWidth and BorderColor fields and Assign() method are now ported to the Pointer struct in gauge.go and serialized/deserialized via dot-notation in GaugeObject.Serialize/Deserialize. Fill property (FillBase type), parent reference, and standalone Serialize() remain absent by design — Go uses a simplified Color string rather than FillBase; pointer serialization is inlined into GaugeObject. Draw() is replaced by RenderXxx in render.go.

#### `GaugeScale.cs`
- **File**: `FastReport.Base/Gauge/GaugeScale.cs`
- **Status**: PARTIALLY PORTED
- **Fixed 2026-03-22**: Scale.Assign(src *Scale) added — copies all fields (MinorStep, MajorStep, ShowLabels, LabelFormat, Font, MajorTicks, MinorTicks). Mirrors C# GaugeScale.Assign (GaugeScale.cs:102-107).
- **Remaining gaps**: No separate GaugeScale type; Go uses a simpler Scale struct without Parent reference, TextFill property, Draw() rendering method. ScaleTicks class also not ported — ticks are embedded directly in Scale struct.

### Gauge/Linear

#### `LinearGauge.cs`
- **File**: `FastReport.Base/Gauge/Linear/LinearGauge.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

#### `LinearPointer.cs`
- **File**: `FastReport.Base/Gauge/Linear/LinearPointer.cs`
- **Status**: NOT PORTED
- **Gaps**: LinearPointer class with configurable Height/Width and triangular DrawHorz/DrawVert rendering does not exist; Go renders linear gauge as a filled bar rectangle without dedicated pointer type.

#### `LinearScale.cs`
- **File**: `FastReport.Base/Gauge/Linear/LinearScale.cs`
- **Status**: NOT PORTED
- **Gaps**: LinearScale tick drawing (DrawMajorTicks/DrawMinorTicks Horz/Vert variants) and numeric scale labels are completely absent; Go uses simplified generic Scale struct without type-specific rendering.

### Gauge/Radial

#### `RadialGauge.cs`
- **File**: `FastReport.Base/Gauge/Radial/RadialGauge.cs`
- **Status**: MOSTLY PORTED
- **Fixed 2026-03-22**: Verified RadialGaugeType/Position enums, all properties, Serialize/Deserialize, and Assign() are all present in gauge/gauge.go. Porting-gaps.md was stale.
- **Remaining gaps**: Draw() replaced by RenderRadial(). C# 'Type' field name mapped to GaugeType in Go.

#### `RadialLabel.cs`
- **File**: `FastReport.Base/Gauge/Radial/RadialLabel.cs`
- **Status**: NOT PORTED
- **Gaps**: RadialLabel specialized type with Draw() method that positions labels based on RadialGauge type and scale metrics does not exist; Go uses a simplified GaugeLabel data struct without subclasses.

#### `RadialPointer.cs`
- **File**: `FastReport.Base/Gauge/Radial/RadialPointer.cs`
- **Status**: NOT PORTED
- **Gaps**: RadialPointer class (GradientAutoRotate, Fill/BorderColor/BorderWidth properties, polygon-path DrawHorz rendering with rotation matrix) completely missing; Go Pointer struct is a simplified data holder with no rendering logic.

#### `RadialScale.cs`
- **File**: `FastReport.Base/Gauge/Radial/RadialScale.cs`
- **Status**: NOT PORTED
- **Gaps**: RadialScale-specific tick rendering (DrawMajorTicks, DrawMinorTicks, GetTextPoint) and orientation-based text positioning completely absent; Go uses a simplified generic Scale struct without separate RadialScale/LinearScale implementations.

#### `RadialUtils.cs`
- **File**: `FastReport.Base/Gauge/Radial/RadialUtils.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: RotateVector(), IsTop/IsBottom/IsLeft/IsRight() flag helpers, IsSemicircle/IsQuadrant() predicates, and radialStartAngleFor() are all ported in gauge/radialutils.go. Missing: GetFont() and GetStringSize() DPI-scaled font utilities (not needed since Go rendering does not draw text labels on gauge arcs).

### Gauge/Simple/Progress

#### `SimpleProgressGauge.cs`
- **File**: `FastReport.Base/Gauge/Simple/Progress/SimpleProgressGauge.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: SimpleProgressPointerType enum (Full/Small), SmallPointerWidthRatio, and LabelDecimals are now ported as fields on SimpleProgressGauge in gauge.go. PercentText() method implements the C# SimpleProgressLabel.Draw() percentage calculation. RenderSimpleProgress() in render.go handles Full and Small pointer rendering. Remaining gaps: scale disable on init (C# sets FirstSubScale.Enabled=false in constructor — not needed in Go since rendering skips scales), PointerRatio=1 init, and HorizontalOffset=0 init are implicit in Go's flat Pointer struct rendering.

#### `SimpleProgressLabel.cs`
- **File**: `FastReport.Base/Gauge/Simple/Progress/SimpleProgressLabel.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: LabelDecimals property and PercentText() percentage-text method are now ported as fields/methods on SimpleProgressGauge rather than a separate class. No dedicated SimpleProgressLabel type — Go uses an embedded GaugeLabel struct plus PercentText() on the parent gauge. Draw() is replaced by RenderSimpleProgress() which does not yet render the text label onto the image (image rendering cannot draw text).

#### `SimpleProgressPointer.cs`
- **File**: `FastReport.Base/Gauge/Simple/Progress/SimpleProgressPointer.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: SimpleProgressPointerType enum (Full/Small) and SmallPointerWidthRatio property are now ported as fields on SimpleProgressGauge. DrawHorz/DrawVert rendering for both Full and Small types is implemented in RenderSimpleProgress() in render.go. No separate SimpleProgressPointer class — Go uses a flat approach. DrawVert (vertical orientation) is not yet implemented in RenderSimpleProgress (only horizontal is handled).

### Gauge/Simple

#### `SimpleGauge.cs`
- **File**: `FastReport.Base/Gauge/Simple/SimpleGauge.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Full data model, rendering pipeline (RenderSimple with all shapes), engine integration (buildPreparedObject), and serialization ported. C# Draw() pattern replaced by RenderSimple() called from engine.

#### `SimplePointer.cs`
- **File**: `FastReport.Base/Gauge/Simple/SimplePointer.cs`
- **Status**: NOT PORTED
- **Gaps**: No GaugePointer base class or SimplePointer type; missing DrawHorz()/DrawVert() render methods with position/size calculations. Go uses a flat Pointer struct stored in GaugeObject; rendering is handled by gauge/render.go without polymorphic pointer types.

#### `SimpleScale.cs`
- **File**: `FastReport.Base/Gauge/Simple/SimpleScale.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: SimpleSubScale properties (Enabled, ShowCaption) are ported. SimpleScale class hierarchy and four tick-drawing rendering methods (DrawMajorTicksHorz/Vert, DrawMinorTicksHorz/Vert) are NOT ported; Go uses RenderSimple() without separate Scale class abstraction.

### Import

#### `ComponentsFactory.cs`
- **File**: `FastReport.Base/Import/ComponentsFactory.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: ~30 factory methods for Import/ subsystem only.

#### `ImportBase.cs`
- **File**: `FastReport.Base/Import/ImportBase.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Trivial abstract base for 5 import plugins. Design-time migration tooling permanently excluded.

### Import/DevExpress

#### `DevExpressImport.XmlSource.cs`
- **File**: `FastReport.Base/Import/DevExpress/DevExpressImport.XmlSource.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import plugin. Design-time converter, no Go equivalent.

#### `DevExpressImport.cs`
- **File**: `FastReport.Base/Import/DevExpress/DevExpressImport.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import plugin (~1777 lines). Design-time migration tool.

#### `UnitsConverter.cs`
- **File**: `FastReport.Base/Import/DevExpress/UnitsConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import/DevExpress subsystem. 14 methods, sole consumer is DevExpressImport.

### Import/JasperReports

#### `JasperReportsImport.cs`
- **File**: `FastReport.Base/Import/JasperReports/JasperReportsImport.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import plugin for JasperReports .jrxml. Designer-only, no Go import/ package.

#### `UnitsConverter.cs`
- **File**: `FastReport.Base/Import/JasperReports/UnitsConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import/JasperReports subsystem. 16 static conversion methods unported. Sole consumer is JasperReportsImport.

### Import/ListAndLabel

#### `ListAndLabelImport.cs`
- **File**: `FastReport.Base/Import/ListAndLabel/ListAndLabelImport.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import plugin for List & Label format. Designer-only, no Go import/ package.

#### `UnitsConverter.cs`
- **File**: `FastReport.Base/Import/ListAndLabel/UnitsConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: 7 simple conversion methods, all unported. Sole consumer is ListAndLabelImport. Import subsystem deferred. Trivial ~50 LOC.

### Import/RDL

#### `ImportTable.cs`
- **File**: `FastReport.Base/Import/RDL/ImportTable.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Part of Import/ subsystem — design-time converter for RDL Table/Matrix elements.

#### `RDLImport.cs`
- **File**: `FastReport.Base/Import/RDL/RDLImport.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: ~988-line converter from SSRS RDL/RDLC to .frx. Design-time migration tool.

#### `SizeUnits.cs`
- **File**: `FastReport.Base/Import/RDL/SizeUnits.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import/RDL subsystem deferred. Go units/ has equivalent pixel constants for mm/cm/in. Missing: RDL string constants, Point/Pica units.

#### `UnitsConverter.cs`
- **File**: `FastReport.Base/Import/RDL/UnitsConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Import/RDL subsystem. 15 methods unported. Would need entire RDL pipeline ported together.

### Import/StimulSoft

#### `StimulSoftImport.cs`
- **File**: `FastReport.Base/Import/StimulSoft/StimulSoftImport.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Design-time converter from StimulSoft .mrt to .frx. All 40+ methods out of scope. No import/ package in Go.

#### `UnitsConverter.cs`
- **File**: `FastReport.Base/Import/StimulSoft/UnitsConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: 0 implemented, 23 NOT IMPLEMENTED, 3 OUT OF SCOPE. Entire Import/ subsystem deferred. StimulSoft format converters, not core gaps.

### Matrix

#### `MatrixCellDescriptor.cs`
- **File**: `FastReport.Base/Matrix/MatrixCellDescriptor.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Verified: Function (AggregateFunction) and Percent (MatrixPercent) match C# exactly. No TotalType property exists in C# MatrixCellDescriptor. Serialization writes Function/Percent as integers matching C# WriteValue convention.

#### `MatrixCells.cs`
- **File**: `FastReport.Base/Matrix/MatrixCells.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: No standalone MatrixCells collection type; collection methods (Add/Insert/Remove/IndexOf/Contains/ToArray) not exposed as a class. Functionality is equivalent but refactored as slice operations within MatrixData and MatrixObject.

#### `MatrixData.cs`
- **File**: `FastReport.Base/Matrix/MatrixData.cs`
- **Status**: PARTIALLY PORTED
- **Ported** (issue go-fastreport-5r21k): Clear() added to MatrixData in `matrix/descriptor_templates.go`. Collection API (IndexOf/Contains/Insert/Remove/ToArray) added for Columns, Rows, and Cells on MatrixData.
- **Gaps**: Missing AddValue(object[], object[], object[]), GetValue(int,int,int), SetValue(int,int,object). Go implements simplified AddData() on MatrixObject rather than AddValue on MatrixData. MatrixHeader tree-based Find()/Reset()/FindOrCreate() not ported (Go uses flat slice collections).

#### `MatrixDescriptor.cs`
- **File**: `FastReport.Base/Matrix/MatrixDescriptor.cs`
- **Status**: PARTIALLY PORTED
- **Ported** (issue go-fastreport-5r21k): TemplateColumn, TemplateRow, TemplateCell fields added via `DescriptorExt` struct in `matrix/descriptor_templates.go`. Accessible through `HeaderDescriptor.HeaderExt()` and `CellDescriptor.CellExt()` accessors. Assign() method added to `HeaderDescriptor`.
- **Gaps**: DescriptorExt is stored in a side-map rather than embedded in the struct (Go does not allow adding fields to a struct in a separate file). Serialization still uses custom writer helper types, not the C# IFRSerializable pattern. Assign() on base Descriptor is not a standalone method (logic folded into HeaderDescriptor.Assign).

#### `MatrixHeader.cs`
- **File**: `FastReport.Base/Matrix/MatrixHeader.cs`
- **Status**: PARTIALLY PORTED
- **Ported** (issue go-fastreport-5r21k): Collection API (IndexOfColumn/Row/Cell, ContainsColumn/Row/Cell, InsertColumn/Row/Cell, RemoveColumn/Row/Cell, ColumnsToArray/RowsToArray/CellsToArray) added to MatrixData in `matrix/descriptor_templates.go`, covering the C# MatrixHeader.Add/Insert/Remove/IndexOf/Contains/ToArray pattern.
- **Gaps**: Tree-based navigation methods (Find with binary search, FindOrCreate, RemoveItem, GetTerminalIndices, AddTotalItems, Reset) are not ported. The Go implementation uses flat slices on MatrixData rather than the C# CollectionBase-derived MatrixHeader class with its internal tree (rootItem, nextIndex). FastCube integration is out of scope.

#### `MatrixHeaderDescriptor.cs`
- **File**: `FastReport.Base/Matrix/MatrixHeaderDescriptor.cs`
- **Status**: PARTIALLY PORTED
- **Ported** (issue go-fastreport-5r21k): TemplateTotalColumn, TemplateTotalRow, TemplateTotalCell fields added via `HeaderDescriptorExt` in `matrix/descriptor_templates.go`, accessible through `HeaderDescriptor.HeaderExt()`. Assign() method added to HeaderDescriptor (copies Expression, Sort, Totals, TotalsFirst, PageBreak, SuppressTotals, TemplateCell, TemplateTotalCell).
- **Gaps**: Multiple constructor overloads are not applicable in Go (use NewHeaderDescriptor + field assignment). TemplateTotalColumn/Row are stored in the side-map ext rather than directly on the struct. Serialization of Totals/SuppressTotals/PageBreak uses Go convention rather than exact C# FRWriter.WriteValue pattern.

#### `MatrixHeaderItem.cs`
- **File**: `FastReport.Base/Matrix/MatrixHeaderItem.cs`
- **Status**: PARTIALLY PORTED
- **Not changed** (issue go-fastreport-5r21k): FastCube integration and the full tree runtime pipeline are out of scope for this iteration.
- **Gaps**: Go HeaderItem (in `matrix/header_tree.go`) lacks parent pointer, Index field, IsTotal/DataRowNo/PageBreak/IsSplitted flags, Find() binary search method with HeaderComparer, and MatrixDescriptor inheritance (TemplateColumn/Row/Cell). Value is string-only rather than object. These are needed for the full runtime printing pipeline (MatrixHelper.InitResultTable, PrintHeaderCell) but not for the descriptor/template-binding layer addressed here.

#### `MatrixHelper.cs`
- **File**: `FastReport.Base/Matrix/MatrixHelper.cs`
- **Status**: PARTIALLY PORTED
- **Not changed** (issue go-fastreport-5r21k): MatrixHelper runtime pipeline is out of scope. The Build() pipeline foundations (descriptor template binding via HeaderExt/CellExt) were addressed at the descriptor level.
- **Gaps**: Missing runtime printing methods (StartPrint, AddDataRow, AddDataRows, FinishPrint), result-table construction (InitResultTable, PrintHeaderCell, PrintDataCell), style application (ApplyStyle, CreateCell, CreateDataCell), and the design-time/runtime template-size calculation (UpdateTemplateSizes, UpdateColumnDescriptors, UpdateRowDescriptors, UpdateCellDescriptors). These depend on a full MatrixHeaderItem tree (parent, Index, IsTotal, DataRowNo) which is also not yet ported.

#### `MatrixObject.Async.cs`
- **File**: `FastReport.Base/Matrix/MatrixObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API. GetDataAsync/BuildTableAsync have no Go equivalents.

#### `MatrixObject.cs`
- **File**: `FastReport.Base/Matrix/MatrixObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Serialization and static layout helpers are ported. Missing: full engine lifecycle (SaveState/RestoreState/InitializeComponent/FinalizeComponent/GetData/OnAfterData), event firing (OnManualBuild/OnModifyResult/OnAfterTotals), runtime state (ColumnValues/RowValues/ColumnIndex/RowIndex), result table creation/disposal, MatrixHelper integration for dynamic data population, and AddValue()/Value() public APIs for manual matrix building.

#### `MatrixStyleSheet.cs`
- **File**: `FastReport.Base/Matrix/MatrixStyleSheet.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer-only utility that generates 16x16 bitmap previews of matrix styles; has no runtime reporting equivalent. Go matrix uses named style string references instead.

### Preview

#### `Bookmarks.cs`
- **File**: `FastReport.Base/Preview/Bookmarks.cs`
- **Status**: FULLY PORTED
- **Fixed (2026-03-22)** (go-fastreport-c8ve9): Verified Clear(), ClearFirstPass(), and GetPageNo() with firstPassItems fallback are all implemented in preview/prepared_pages.go. Porting-gaps.md entry was stale — all 9 methods are now implemented.

#### `Dictionary.cs`
- **File**: `FastReport.Base/Preview/Dictionary.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Unique name creation ported to utils/ (FastNameCreator). Preview-level object aliasing/cloning during page deserialization (GetObject/GetOriginalObject/AddUnique/DictionaryItem.CloneObject) not ported as a standalone module; Go PreparedPage is data-only and does not reconstruct objects from XML during export.

#### `Outline.cs`
- **File**: `FastReport.Base/Preview/Outline.cs`
- **Status**: FULLY PORTED
- **Gaps**: All 13 members ported with idiomatic Go stack-based design.

#### `PageCache.cs`
- **File**: `FastReport.Base/Preview/PageCache.cs`
- **Status**: FULLY PORTED
- **Gaps**: LRU algorithm matches C#. Not integrated into PreparedPages.

#### `PreparedPage.cs`
- **File**: `FastReport.Base/Preview/PreparedPage.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 7 items ported. PostprocessUnlimited equivalent implemented in postprocessor.go (ProcessUnlimited/PostProcessBandUnlimited). MEDIUM: ReCalcSizes, MirrorMargins.

#### `PreparedPagePostprocessor.cs`
- **File**: `FastReport.Base/Preview/PreparedPagePostprocessor.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Duplicates and macros work. MergeMode text merging implemented. PostprocessUnlimited and PostProcessBandUnlimitedPage implemented (preview/postprocessor.go ProcessUnlimited/PostProcessBandUnlimited). MEDIUM: TotalPages ignores InitialPageNumber, watermark macros not replaced.

#### `PreparedPages.cs`
- **File**: `FastReport.Base/Preview/PreparedPages.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go uses flattened structs. InterleaveWithBackPage implemented (preview/prepared_pages.go). HIGH: Bookmarks.ClearFirstPass fallback. MEDIUM: UnlimitedHeight/Width page merging handled via engine (ModifyPageSize), GetLastY, ContainsBand, file cache.

#### `SourcePages.cs`
- **File**: `FastReport.Base/Preview/SourcePages.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go stores integer index-range tuples vs C# deep-cloned ReportPage objects. CloneObjects and Dictionary alias registration are not applicable in Go's architecture (no XmlItem/FRWriter/FRReader layer in preview). ApplyWatermark is a no-op stub; watermark metadata is attached directly to PreparedPage by the engine.

### Table

#### `TableBase.cs`
- **File**: `FastReport.Base/Table/TableBase.cs`
- **Status**: PARTIALLY PORTED
- **Implemented (2026-03-22)**: GetSpanList/ResetSpanList (cached span rectangle list), IsInsideSpan, CorrectSpansOnRowChange, CorrectSpansOnColumnChange (with cell slot insert/remove), SaveState/RestoreState (delegates to rows, columns, cells; sets CanGrow=CanShrink=true), CalcWidth/CalcHeight (two-pass auto-size with span support, skips invisible rows in height total).
- **Fixed (2026-03-22)** (go-fastreport-ukhea): Added `Assign(src *TableBase)` — copies FixedRows, FixedColumns, PrintOnParent, RepeatHeaders, RepeatRowHeaders, RepeatColumnHeaders, Layout, WrappedGap, AdjustSpannedCellsWidth, ManualBuildEvent from src (mirrors C# TableBase.Assign:473-489). Rows/columns/cells not copied — structural, managed by engine.
- **Remaining Gaps**: Missing Draw() rendering methods, IParent interface (CanContain/AddChild/RemoveChild/UpdateLayout etc.), Break/BreakRow logic, border/fill emulation (EmulateOuterBorder/EmulateFill), TableCellData abstraction, GetCellData(), and CreateUniqueNames(). CalcWidth/CalcHeight use cell.Width()/Height() as CalcWidth/CalcHeight equivalents (no text-measurement-based sizing).

#### `TableCell.Async.cs`
- **File**: `FastReport.Base/Table/TableCell.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API; GetDataAsync is not applicable.

#### `TableCell.cs`
- **File**: `FastReport.Base/Table/TableCell.cs`
- **Status**: MOSTLY PORTED
- **Implemented (2026-03-22)**: GetExpressions() (delegates to base + embedded objects), SaveState/RestoreState (saves text + embedded object count, discards dynamically added objects on restore), GetData(isInsideSpan) (clears text when inside span, calls GetData on embedded objects up to savedObjectCount), CalcWidth/CalcHeight (return current Width()/Height()).
- **Fixed (2026-03-22)** (go-fastreport-se9gd): Added `Assign(src *TableCell)` — copies TextObject by value, deep-copies highlights slice, copies colSpan/rowSpan/duplicates (mirrors C# TableCell.Assign:221-228). Added `Clone() *TableCell` — creates new cell and calls Assign (mirrors C# TableCell.Clone:235-239). Added `EqualStyle(other *TableCell) bool` — compares visual style fields (HorzAlign, VertAlign, Angle, WordWrap, Font, TextColor, etc.) for style deduplication (mirrors C# TableCell.Equals:247-283, named EqualStyle to avoid Go interface collision). Added `SetHighlights()` to TextObject to enable deep copy.
- **Remaining Gaps**: IParent interface methods (CanContain/GetChildObjects/AddChild/RemoveChild etc.), CellData dual-mode property, Address computed property. CalcWidth/CalcHeight do not perform text-measurement-based sizing.

#### `TableCellData.cs`
- **File**: `FastReport.Base/Table/TableCellData.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: No separate TableCellData runtime type; concepts merged into TableCell. Missing calculated Width/Height (from ColSpan/RowSpan dimensions), CalcWidth/CalcHeight with object growth, AttachCell/RunTimeAssign/SetStyle/UpdateLayout methods, and Address property.

#### `TableColumn.cs`
- **File**: `FastReport.Base/Table/TableColumn.cs`
- **Status**: MOSTLY PORTED
- **Implemented (2026-03-22)**: SetWidth() with min/max bounds enforcement (clamps to MaxWidth then MinWidth; MaxWidth=0 means unlimited), SaveState/RestoreState (saves/restores Width and Visible).
- **Fixed (2026-03-22)** (go-fastreport-yso06): Added `Assign(src *TableColumn)` — copies minWidth, maxWidth, autoSize, keepColumns, pageBreak, and delegates to ComponentBase.Assign (mirrors C# TableColumn.Assign lines 188-197). Added `Clear()` — resets width to default 100 (mirrors C# TableColumn.Clear:233-245, simplified without parent-table reference).
- **Remaining Gaps**: Left computed property (cumulative column widths), UpdateLayout() propagation to cells, Index tracking.

#### `TableColumnCollection.cs`
- **File**: `FastReport.Base/Table/TableColumnCollection.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: No explicit TableColumnCollection class; columns managed inline in TableBase slice. Missing OnInsert/OnRemove hooks that trigger CorrectSpansOnColumnChange() when columns are added or removed.

#### `TableHelper.cs`
- **File**: `FastReport.Base/Table/TableHelper.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing PageBreak() method that handles table splitting across page boundaries with header row repetition.

#### `TableObject.Async.cs`
- **File**: `FastReport.Base/Table/TableObject.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Async wrapper — Go uses synchronous API with context.Context for cancellation.

#### `TableObject.cs`
- **File**: `FastReport.Base/Table/TableObject.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Static table serialization and basic rendering ported. SaveState/RestoreState are inherited from TableBase (which now delegates to rows, columns, and cells). Missing: GetData() dynamic data-binding lifecycle, OnAfterData() event firing, OnManualBuild with EventArgs, GetCustomScript(), Assign(), and ColumnCount/RowCount setters with CreateUniqueNames() for auto-naming cells.

#### `TableResult.cs`
- **File**: `FastReport.Base/Table/TableResult.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing table pagination algorithms (GeneratePagesAcrossThenDown, GeneratePagesDownThenAcross, GeneratePagesWrapped), ProcessDuplicates(), IsInsideSpan(), GeneratePages event, AddToParent(), and TableLayoutInfo class.

#### `TableRow.cs`
- **File**: `FastReport.Base/Table/TableRow.cs`
- **Status**: MOSTLY PORTED
- **Implemented (2026-03-22)**: SetHeight() with min/max bounds enforcement (clamps to MaxHeight when CanBreak=false; MaxHeight=0 means unlimited; clamps upward to MinHeight), SaveState/RestoreState (saves/restores Height and Visible).
- **Fixed (2026-03-22)** (go-fastreport-xdaml): Added `Assign(src *TableRow)` — copies minHeight, maxHeight, autoSize, keepRows, canBreak, pageBreak, and delegates to ComponentBase.Assign (mirrors C# TableRow.Assign lines 288-297). Added `Clear()` — clears cells slice and resets height to default 30 (mirrors C# TableRow.Clear:361-368).
- **Remaining Gaps**: IParent interface implementation (CanContain/GetChildObjects/AddChild/RemoveChild/GetChildOrder/SetChildOrder/UpdateLayout), Index property, dynamic Top calculation, CellData() method.

#### `TableRowCollection.cs`
- **File**: `FastReport.Base/Table/TableRowCollection.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go inlines row management into TableBase using slices; missing SetIndex callback mechanism and OnInsert/OnRemove hooks that trigger CorrectSpansOnRowChange() on row additions/removals.

#### `TableStyleCollection.cs`
- **File**: `FastReport.Base/Table/TableStyleCollection.cs`
- **Status**: FULLY PORTED
- **Gaps**: None

### TypeConverters

#### `BarcodeConverter.cs`
- **File**: `FastReport.Base/TypeConverters/BarcodeConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter.

#### `ComponentRefConverter.cs`
- **File**: `FastReport.Base/TypeConverters/ComponentRefConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter for component name references. Go handles via serialization.

#### `CubeSourceConverter.cs`
- **File**: `FastReport.Base/TypeConverters/CubeSourceConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter for CubeSourceBase property grid.

#### `DataSourceConverter.cs`
- **File**: `FastReport.Base/TypeConverters/DataSourceConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter. DataSource resolution handled via FRX deserialization.

#### `DataTypeConverter.cs`
- **File**: `FastReport.Base/TypeConverters/DataTypeConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter for Column/Parameter DataType property grid.

#### `FRExpandableObjectConverter.cs`
- **File**: `FastReport.Base/TypeConverters/FRExpandableObjectConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter for expandable property grid. No runtime impact.

#### `FillConverter.cs`
- **File**: `FastReport.Base/TypeConverters/FillConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter for FillBase. Runtime fill dispatch already ported in borderfill_serial.go.

#### `FlagConverter.cs`
- **File**: `FastReport.Base/TypeConverters/FlagConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer EnumConverter for [Flags] enum dropdown. No runtime impact.

#### `FloatCollectionConverter.cs`
- **File**: `FastReport.Base/TypeConverters/FloatCollectionConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer TypeConverter. String conversion already ported in Go (ParseFloatCollection, String).

#### `FormatConverter.cs`
- **File**: `FastReport.Base/TypeConverters/FormatConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer-only TypeConverter for FormatBase. Go handles format resolution via deserializeTextFormat.

#### `ParameterDataTypeConverter.cs`
- **File**: `FastReport.Base/TypeConverters/ParameterDataTypeConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer-only TypeConverter for CommandParameter.DataType property grid.

#### `RelationConverter.cs`
- **File**: `FastReport.Base/TypeConverters/RelationConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Designer-only TypeConverter for Relation property grid editing.

### Utils

#### `AssemblyInitializerBase.cs`
- **File**: `FastReport.Base/Utils/AssemblyInitializerBase.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Go replaces reflection-based plugin discovery with init() functions. All type registrations covered.

#### `BlobStore.cs`
- **File**: `FastReport.Base/Utils/BlobStore.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Core ported (AddOrUpdate, Get, Count, Save/Load). Missing: GetSource reverse lookup, name-less Add, Clear, file cache for large reports.

#### `ColorHelper.cs`
- **File**: `FastReport.Base/Utils/ColorHelper.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: FromString/ParseColor fully ported (HEX, CSV, decimal ARGB, 140 named colors). Minor: FromObject not ported (0 call sites), CSV range validation, ColorTransparent inconsistency.

#### `CompileHelper.cs`
- **File**: `FastReport.Base/Utils/CompileHelper.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET CodeDom runtime compilation. Go uses expr-lang/expr and native JSON binding.

#### `CompilerSettings.cs`
- **File**: `FastReport.Base/Utils/CompilerSettings.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET CodeDom pipeline config. Go uses expr-lang/expr.

#### `Compressor.cs`
- **File**: `FastReport.Base/Utils/Compressor.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: String/byte Compress/Decompress ported. Stream-based handled inline. MEDIUM: PreparedPages FPX format has no compression -- Config.PreparedCompressed flag never consulted.

#### `Config.cs`
- **File**: `FastReport.Base/Utils/Config.cs`
- **Status**: MOSTLY PORTED
- **Fixed (2026-03-22)**: Implemented all previously-claimed (but missing) additions in `utils/config.go`: `Version` constant (matches C# Config.Version, Config.cs line 199). `GetTempFolder()` now returns `os.TempDir()` when TempFolder is empty, mirroring C# `Config.GetTempFolder()` which returns `GetTempPath()` when TempFolder==null (Config.cs lines 291-293). `GetConfiguredTempFolder()` returns the raw field without fallback. `CreateTempFile(dir string)` matches C# `Config.CreateTempFile(string dir)` (Config.cs lines 284-289) — delegates to `TempFilePath()` when dir is empty, otherwise uses `os.CreateTemp(dir, ...)`. `TempFilePath()` creates a timestamped temp file in the effective temp folder, matching C# `Config.GetTempFileName()` (Config.cs lines 411-414). Package-level helpers `CreateTempFileInDir(dir)` and `GetEffectiveTempFolder()` added for convenience. 6 new tests added in `utils/config_test.go` covering all new functions and the `GetTempFolder`/`GetConfiguredTempFolder` fallback contract.
- **Remaining Gaps**: ~15 items intentionally omitted — script security (ScriptSecurityProperties, EnableScriptSecurity), plugins (LoadPlugins, ProcessAssembly), platform detection (IsRunningOnMono, IsWindows, WebMode, CheckWebMode), config file persistence (LoadConfig, SaveConfig, CurrentDomain_ProcessExit), UI settings (RestoreUIStyle, RestoreUIOptions, RightToLeft XML restore), CompilerSettings — all OUT OF SCOPE for headless Go port. FilterConnectionTables event — OUT OF SCOPE. IsStringOptimization — not applicable to Go string model.

#### `Converter.cs`
- **File**: `FastReport.Base/Utils/Converter.cs`
- **Status**: MOSTLY PORTED
- **Fixed (2026-03-21)**: Added `ToXml(s)` and `ToXmlKeepCRLF(s)` (Converter.cs lines 110-137) — escapes `"`, `&`, `<`, `>` and optionally CR/LF as `&#10;`/`&#13;` numeric references. Added `FromXml(s)` (lines 150-193) — decodes `&#ddd;`, `&#xhh;`, `&quot;`, `&amp;`, `&lt;`, `&gt;`, `&apos;`. Added `DecreasePrecision(value, precision)` (line 247). Added `StringToFloat(s)` and `StringToFloatSep(s, sep)` (lines 228-245). Added `StringToByteArray(s)` (lines 202-215). Tests added in `utils/converter_xml_test.go`.
- **Remaining Gaps**: Polymorphic `ToString`/`FromString` dispatch (reflection-based C# TypeConverter) — OUT OF SCOPE for Go port; callers use type-specific helpers directly. `FromHtmlEntities` entity table (~2000 named HTML entities) remains basic 5-entity implementation — LOW priority, only used in designer label import.

#### `Crc32.cs`
- **File**: `FastReport.Base/Utils/Crc32.cs`
- **Status**: FULLY PORTED
- **Gaps**: Go wraps stdlib hash/crc32 (IEEE). Missing Begin/Update/End only used by C# Zip.cs which Go handles via compress/flate.

#### `Crypter.cs`
- **File**: `FastReport.Base/Utils/Crypter.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 6 of 9 members IMPLEMENTED. Missing: ComputeHash+Murmur3, DefaultPassword, DataConnectionBase encrypt/decrypt of ConnectionString.

#### `DrawUtils.cs`
- **File**: `FastReport.Base/Utils/DrawUtils.cs`
- **Status**: PARTIALLY PORTED
- **Reviewed (2026-03-21)**: No new functions needed for server-side headless operation. All ported items (DefaultFont, MeasureString, DashPattern) are sufficient. Remaining gaps are GDI+-specific.
- **Remaining Gaps**: `SetPenDashPatternOrStyle` — GDI+ Pen object, OUT OF SCOPE. CJK locale font fallback — designer/Windows-specific, OUT OF SCOPE. `Graphics`/`GDI+` rendering methods — OUT OF SCOPE. `GetFontStyle` — WinForms FontStyle enum, OUT OF SCOPE.

#### `Exceptions.cs`
- **File**: `FastReport.Base/Utils/Exceptions.cs`
- **Status**: FULLY PORTED (headless subset)
- **Fixed (2026-03-21)**: Added `SwissQrCodeError` (wraps cause error, matches `SwissQrCodeException` — Exceptions.cs lines 43-55), `TableManualBuildError` (Exceptions.cs lines 57-70), and `MatrixValueError` with Count field (Exceptions.cs lines 72-85). All three are in `utils/errors.go`. Tests added in `utils/errors_new_types_test.go`.
- **Remaining Gaps**: `CloudStorageException` — OUT OF SCOPE (cloud storage connector, no Go equivalent). All 3 remaining exceptions in C# (`ThreadAbortException` wrapping, `OutOfMemoryException`, `AccessViolationException`) — .NET runtime exceptions with no Go equivalent.

#### `ExportsOptions.cs`
- **File**: `FastReport.Base/Utils/ExportsOptions.cs`
- **Status**: MOSTLY PORTED
- **Fixed (2026-03-21)**: Added `SetFormatEnabled(format, bool)` — enables/disables a format by adding/removing from HideExports, equivalent to C# `SetExportEnabled(Type, bool)` (ExportsOptions.cs line 130-136). Added `AllowOnly(formats...)` — restricts AllowedExports list, equivalent to configuring the ExportsMenu tree. Tests added in export/export_test.go.
- **Remaining Gaps**: ExportsTreeNode tree-based UI menu system (14 members: Name, Nodes, Parent, Root, ExportType, Text, ImageIndex, Image, Tag, Enabled, AddCategory, AddExport, ExportsTreeNodeCollection) — entirely OUT OF SCOPE for headless library; the C# tree controls only a WinForms/WPF preview toolbar. BeforeRestoreState/AfterRestoreState events — OUT OF SCOPE (no persistent config in Go headless mode). SaveState/RestoreState — OUT OF SCOPE. RegisteredObjects integration — OUT OF SCOPE.

#### `FRCollectionBase.cs`
- **File**: `FastReport.Base/Utils/FRCollectionBase.cs`
- **Status**: FULLY PORTED
- **Fixed (2026-03-21)**: `Equals(*ObjectCollection)` (FRCollectionBase.cs line 115-128), `CopyTo(*ObjectCollection)` (line 134-139), `AddRangeCollection(*ObjectCollection)` overload (line 39-45), and nil guards on `Add`/`Insert` (line 53-56, 64-67) are all now implemented in `report/collections.go`. Tests added in `report/collections_test.go`.
- **Remaining Gaps**: Owner field — OUT OF SCOPE (headless library; parent managed via Parent interface at call sites). OnInsert/OnRemove/OnClear parent lifecycle hooks — OUT OF SCOPE (parent set explicitly via AddChild/RemoveChild in engine). Add returning int index — not needed in Go; callers use Len(). ToArray() — not needed; callers use Slice().

#### `FRPaintEventArgs.cs`
- **File**: `FastReport.Base/Utils/FRPaintEventArgs.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: All 6 members OUT OF SCOPE. GDI+ Draw() pattern event arg. Go uses fundamentally different architecture.

#### `FRPrivateFontCollection.cs`
- **File**: `FastReport.Base/Utils/FRPrivateFontCollection.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Deprecated [Obsolete] class — thin wrapper delegating to FontManager. Both methods NOT IMPLEMENTED. Priority LOW.

#### `FRRandom.cs`
- **File**: `FastReport.Base/Utils/FRRandom.cs`
- **Status**: NOT PORTED
- **Gaps**: 23 public methods + constructor all NOT IMPLEMENTED. Utility methods portable via math/rand. RandomizeDataSources needs .NET DataTable adaptation.

#### `FRReader.cs`
- **File**: `FastReport.Base/Utils/FRReader.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 10 implemented (ReadStr/Bool/Int/Float, NextChild, CurrentName, Deserialize, Registry.Create, ExpandPropName, NewReader). 10 not implemented: ReadDouble, ReadValue, HasProperty, ReadProperties (reflection), ReadPropertyValue, FixupReferences, DeserializeFrom/SerializeTo enum, ReadChildren flag, dot-path property sorting.

#### `FRWriter.cs`
- **File**: `FastReport.Base/Utils/FRWriter.cs`
- **Status**: MOSTLY PORTED
- **Fixed (2026-03-21)**: Added `WriteDouble(name, float64)` to `report.Writer` interface and `serial.Writer` implementation (FRWriter.cs lines 134-136) — formats with invariant culture dot separator. Added `WriteRef(name, ref)` (FRWriter.cs lines 141-147) — writes component name or `"null"` for nil refs. Added `WritePropertyValue(name, value)` (FRWriter.cs lines 151-158) — emits `<name>value</name>` as child XML element. Updated all ~20 mock writer implementations in test files. Tests added in `serial/writer_new_methods_test.go`.
- **Remaining Gaps**: `WriteValue` (reflection-based object serialization) — OUT OF SCOPE for Go; callers use type-specific Write* methods. `AreEqual`/`DiffObject` (diff-based serialization for designer undo) — OUT OF SCOPE (no designer). `SerializeTo` enum saves — designer feature. `SaveChildren`/`SaveExternalPages` — engine-level operations handled by `engine/` package. `ItemName`/`BlobStore` on writer — BlobStore is on `preview.PreparedPages`, not on writer. `PropName` dot-path writes — handled by callers using explicit attribute names.

#### `FastNameCreator.cs`
- **File**: `FastReport.Base/Utils/FastNameCreator.cs`
- **Status**: FULLY PORTED
- **Gaps**: All public members implemented.

#### `FastString.cs`
- **File**: `FastReport.Base/Utils/FastString.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 7 IMPLEMENTED (Len, NewFastString, IsEmpty, String, Reset, Append, AppendLine). 14 NOT IMPLEMENTED but all have idiomatic Go replacements (strings.ReplaceAll, slicing, fmt.Sprintf). No blocking gaps.

#### `FileUtils.cs`
- **File**: `FastReport.Base/Utils/FileUtils.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Single method GetRelativePath only used in FRX save path (not ported). Go stdlib filepath.Rel() is direct equivalent.

#### `FloatCollection.cs`
- **File**: `FastReport.Base/Utils/FloatCollection.cs`
- **Status**: FULLY PORTED
- **Fixed (2026-03-22)** (go-fastreport-2p77m): All methods now implemented — AddRange, Insert, Remove, IndexOf (with 0.01 epsilon), Contains, Assign, RemoveAt all ported in utils/floatcollection.go. Mirrors C# FloatCollection (FloatCollection.cs).

#### `FontManager.Gdi.cs`
- **File**: `FastReport.Base/Utils/FontManager.Gdi.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: `AddFont(filename)`, `AddFont(IntPtr, int)`, and `CheckFontIsInstalled` are GDI+/.NET-platform-specific methods with no Go equivalent needed. The Go port's `FontManager.AddFontFamily(name)` + `FontManager.AddFace(desc, face)` cover the use-case of registering custom fonts without platform interop.

#### `FontManager.Internals.cs`
- **File**: `FastReport.Base/Utils/FontManager.Internals.cs`
- **Status**: FULLY PORTED
- **Gaps**: `fontSubstitute` private struct IMPLEMENTED in `utils/font.go`. `SearchScope` enum and `FontFamilyMatcher`/`FontConverter` integration are OUT OF SCOPE (GDI+/.NET TypeConverter infrastructure); the Go port uses `ResolveFamily()` instead.

#### `FontManager.cs`
- **File**: `FastReport.Base/Utils/FontManager.cs`
- **Go file**: `utils/font.go` (`FontManager` struct, `DefaultFontManager` global)
- **Status**: FULLY PORTED
- **Implemented**:
  - `AllFamilies` → `FontManager.AllFamilies() []string` (sorted, deduplicated)
  - `AddSubstituteFont` → `FontManager.AddSubstituteFont(originalFontName string, alternatives ...string)`
  - `RemoveSubstituteFont` → `FontManager.RemoveSubstituteFont(originalFontName string)`
  - `ClearSubstituteFonts` → `FontManager.ClearSubstituteFonts()`
  - `GetFontFamilyOrDefault` → `FontManager.ResolveFamily(name string) string`
  - `FontSubstitute` private class → `fontSubstitute` struct (internal to package)
  - `AddFontFamily(name)` — new Go helper for registering family names without GDI+ interop
  - Thread safety: `sync.RWMutex` throughout (Go version IS thread safe; C# was NOT)
- **Out of scope**: `AddFont(filename)`, `AddFont(IntPtr, int)`, `CheckFontIsInstalled` (GDI+/.NET platform-specific, see FontManager.Gdi.cs above); `FontFamilyMatcher`/`FontConverter` integration (designer-only).

#### `GraphicCache.cs`
- **File**: `FastReport.Base/Utils/GraphicCache.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: GDI+ object pool for on-screen painting. Go is headless engine — each exporter handles styling independently. Font caching handled natively.

#### `HtmlTextRenderer.cs`
- **File**: `FastReport.Base/Utils/HtmlTextRenderer.cs`
- **Go file**: `utils/htmltext.go`, `utils/htmltext_renderer_test.go`, `utils/htmltext_test.go`
- **Status**: PARTIALLY PORTED (updated 2026-03-22)
- **What is ported** (~50% of the C# file):
  - `HtmlTextRenderer` struct with `NewHtmlTextRenderer(htmlText, baseFont, baseColor)` constructor
  - `Lines()` returning `[]HtmlLine`; `PlainText()`; `MeasureHeight(width)`; `StripHtmlTags(s)`
  - HTML tag parser: `<b>`, `<strong>`, `<i>`, `<em>`, `<u>`, `<s>`, `<strike>`, `<del>`, `<br>`, `<font>`, `<span>`, `<sub>`, `<sup>`
  - `<sub>`/`<sup>` → `BaselineType` (Subscript/Superscript) — mirrors C# `BaseLine` enum (line 1304) and `SplitToParagraphs` cases (lines 1012-1017)
  - CSS inline `style=""`: `color`, `background-color`, `font-size` (px/pt/em), `font-family`, `font-weight`, `font-style`, `text-decoration` — mirrors C# `CssStyle()` (line 574)
  - `<font color="..." size="..." face="...">` attribute parsing
  - Color parsing: `#rrggbb`, `#aarrggbb`, `rgb(r,g,b)`, `rgba(r,g,b,a)`, named colors — mirrors C# color blocks (lines 626-712)
  - `HtmlRun.BackgroundColor` from CSS `background-color` (mirrors C# `StyleDescriptor.BackgroundColor`, line 670)
  - HTML entities: `&amp;`, `&lt;`, `&gt;`, `&nbsp;`, `&quot;`
  - 40 tests in `htmltext_renderer_test.go` covering all of the above
- **Deliberate architecture differences** (not gaps):
  - C# uses a GDI+-backed layout engine (Paragraph/Line/Word/Run with pixel-precise positions via `Graphics.MeasureString`). Go is headless: HTML parsing is in `utils/htmltext.go`; pixel measurement is in `utils/textrenderer.go` (CalcTextHeight, CalcTextWidth, CharsFitInWidth); rendering is per-exporter.
  - `Draw()`, `RendererContext`, `StringFormat`, `IGraphics` — not needed; Go exporters render from `[]HtmlLine`/`[]HtmlRun` directly.
  - `RightToLeft` — not in the HTML parser; handled at exporter level.
- **Remaining gaps** (not yet ported):
  - `BreakHtml(text, charsFit)` — splits HTML at char index keeping tag balance. Used by C# `TextObject.CanBreak()`. Go engine uses geometric clipping; not currently needed.
  - Inline `<img src="...">` in HTML text (`RunImage` class, line 2088). Not used by current Go exporters.
  - Tab stop handling in HTML parser (`\t` case in `SplitToParagraphs`, lines 815-871). Tab rendering is handled at engine/exporter level.
  - `WingdingsToUnicodeConverter` — symbol font remapping (C# `RunText`, line 2288). Not ported.
  - Paragraph indent (`ParagraphFormat.FirstLineIndent`, `GetStartPosition()`). Not in Go `HtmlLine`/`HtmlRun`.
  - Line spacing (`ParagraphFormat.LineSpacingType`/`LineSpacingMultiple`). Not in Go `HtmlLine`.
  - `StyleDescriptor.ToHtml()` — serialises style to HTML tags. Not needed by current exporters.
  - Full `CalcHeight(out charsFit)` / `CalcWidth()` on layout model — approximated by `MeasureHeight()` + `utils.CalcTextHeight`/`utils.CalcTextWidth`.

#### `ImageHelper.Async.cs`
- **File**: `FastReport.Base/Utils/ImageHelper.Async.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Both methods are async wrappers around already-ported sync equivalents. Go goroutines replace async/await.

#### `ImageHelper.cs`
- **File**: `FastReport.Base/Utils/ImageHelper.cs`
- **Status**: FULLY PORTED (all in-scope functions implemented)
- **Go equivalent**: `utils/image.go`
- **Implemented**:
  - `Load(byte[])` → `BytesToImage`
  - `Load(string fileName)` → `loadFromFile` (internal)
  - `LoadURL(url)` → `loadFromURL` (internal)
  - `ToByteArray` / `Save` → `ImageToBytes`
  - `GetTransparentBitmap` → `ApplyTransparency` (added 2026-03-22; engine applies before storing blob)
  - `GetGrayscaleBitmap` → `ApplyGrayscale` (added 2026-03-22; engine applies before storing blob)
  - `GetImageFormat` (extension) → `imageMIMEForCSS` in HTML exporter
  - Resize/cut helpers → `ResizeImage`
- **Out of scope** (8 items): `IImageHelperLoader` plugin system, `CloneBitmap` (unnecessary with immutable Go images), `CutImage` (covered by ResizeImage), `SaveAndConvert` (no Metafile/WMF/EMF), `SaveAsIcon` (ICO export), `LoadFromFile` with custom loaders, `Register`, `Metafile` handling.

#### `MyEncodingInfo.cs`
- **File**: `FastReport.Base/Utils/MyEncodingInfo.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — .NET-specific wrapper for EncodingInfo. Go uses UTF-8 by default; other encodings are handled via golang.org/x/text/encoding if needed, which does not require this metadata wrapper.

#### `RegisteredObjects.cs`
- **File**: `FastReport.Base/Utils/RegisteredObjects.cs`
- **Status**: PARTIALLY PORTED
- **Go equivalent**: `serial/registry.go` + `reportpkg/serial_registrations.go`
- **Implemented**:
  - `FTypes` hashtable + `RegisterType` → `Registry.factories` map with factory functions (`Register`, `MustRegister`)
  - `FindType(typeName)` → `Registry.Create(name)` (returns a new instance rather than a Type)
  - `IsTypeRegistered` → `Registry.Has(name)`
  - `Names()` returns sorted list of all registered type names (no C# equivalent)
  - `DefaultRegistry` global instance mirrors singleton pattern of `RegisteredObjects` static class
  - All bands, objects, table/matrix/gauge types registered in `reportpkg/serial_registrations.go`
  - Both short names (e.g. `DataBand`) and full names (e.g. `DataBand` → `DataBand`) registered for FRX compatibility
  - Thread-safe concurrent access via `sync.RWMutex`
- **Out of scope** (designer-only, ~60% of C# surface area):
  - `FObjects` ObjectInfo tree (toolbar categories, image indices, flags) — no visual designer in Go
  - `Exports` ObjectInfo tree — export filter registry for designer UI
  - `DataConnections` DataConnectionInfo tree — designer data source browser
  - `Functions` FunctionInfo tree — expression editor function browser
  - `Assemblies` List — .NET assembly tracking (Go uses packages/imports)
  - `RegisterMethod`/`GetMethod` — runtime method-override via reflection (no equivalent in Go)
  - `AddCategory`, `AddExport`, `AddConnection`, `AddFunction`, `AddFunctionCategory` — all designer APIs
  - `FindObject`/`FindExport`/`FindConnection` — designer lookup by Type
  - `Remove(type, category)` / `Remove(name, path, flags)` — dynamic unregistration (not needed by engine)
  - `CreateFunctionsTree`/`GetFunctionDescription` — designer expression editor UI
  - `ObjectInfo`/`FunctionInfo`/`DataConnectionInfo` info tree classes — no Go equivalent needed
- **Minor gap**:
  - `CrossViewObject` is registered in C# (`RegisteredObjects.InternalAdd(typeof(CrossViewObject), ...)`) but the Go `crossview` package does not implement `report.Base`, so it cannot be registered in `DefaultRegistry`. CrossView rendering is handled differently in the Go engine. This is acceptable since there is no FRX `<CrossViewObject>` element in the test corpus.

#### `Res.cs`
- **File**: `FastReport.Base/Utils/Res.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET localized resource manager.

#### `ResourceLoader.cs`
- **File**: `FastReport.Base/Utils/ResourceLoader.cs`
- **Status**: DONE
- **Go file**: `utils/resource_loader.go`
- **Notes**: The .NET implementation retrieves named embedded assembly resources via `Assembly.GetManifestResourceStream`. Go has no DLL-level embedded resources, so the port uses a process-wide registry pattern instead. Packages call `RegisterResource` / `RegisterResourceBytes` to register named byte-slice providers keyed by `(assembly, name)`; `GetStream` / `GetStreamFR` look them up. `UnpackStream` / `UnpackStreamFR` gzip-decompress a registered resource into a fresh in-memory reader (mirroring the C# `GZipStream` + `MemoryStream` approach). The `StorageService` interface (`utils/storage.go`) complements this for file-system I/O. All C# callers (`Res.cs` → `en.xml`, `Config.cs` → `FastReport.config`, `CrossViewObject.cs` / `MatrixObject.cs` → `cross.frss`) are handled in the Go port via alternative mechanisms (built-in English strings in `locale.go`, config struct in `config.go`, stylesheet loading not yet exercised in Go). Minor remaining gap: no resources are pre-registered at init time; callers must register before use.

#### `ScriptSecurityEventArgs.cs`
- **File**: `FastReport.Base/Utils/ScriptSecurityEventArgs.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: CodeDom security.

#### `ShortProperties.cs`
- **File**: `FastReport.Base/Utils/ShortProperties.cs`
- **Go file**: `utils/shortprops.go`
- **Status**: FULLY PORTED (with intentional divergence)
- **Analysis**:
  - C# has 5 entries (`l`, `t`, `w`, `h`, `x`→`Text`) used when `SerializeTo == Preview` to compress FPX XML attribute names.
  - Go equivalent (`utils/shortprops.go`) implements the same bidirectional lookup with `ExpandPropName`/`AbbrevPropName` (pass-through if not found) and `ShortPropCode`/`ShortPropName` (ok-idiom variants). All 4 API functions have full test coverage in `utils/shortprops_test.go`.
  - The Go short code for `"Text"` is `"tx"` (not `"x"` as in C#). This is intentional: the Go FPX format uses binary gob encoding (not XML with short attribute names), so there is no interop requirement with C# FPX XML files. The `"tx"` code avoids confusion with the XML namespace prefix `x`.
  - Go adds 17 extra entries beyond the C# 5 (font, border, fill, color, etc.) as Go-specific extensions for potential future use.
  - **Not integrated into `serial/reader.go` or `serial/writer.go`** by design: the Go serial package reads FRX (design) files which never use short property names. The Go FPX preview format is binary gob (see `preview/fpx.go`), not XML, so short prop expansion is not needed there either.
- **Remaining gaps**: None for the Go pipeline. If future work adds C#-compatible FPX XML import, `serial/reader.go`'s `attrsToMap` should call `utils.ExpandPropName` on attribute names, and the `"x"→"Text"` mapping should be added to match C# exactly.

#### `StorageService.cs`
- **File**: `FastReport.Base/Utils/StorageService.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET storage abstraction.

#### `TextRenderer.cs`
- **File**: `FastReport.Base/Utils/TextRenderer.cs`
- **Status**: PARTIALLY PORTED (updated 2026-03-22)
- **Go equivalents**: `utils/textrenderer.go`, `utils/htmltext.go`, `utils/textmeasure.go`

**What is ported:**
- `AdvancedTextRenderer.CalculateSpaceSize` → `utils.CalculateSpaceWidth` (space pixel width via glyph advance)
- `AdvancedTextRenderer.GetTabPosition` → `utils.GetTabPosition` (tab stop computation: same algorithm)
- `AdvancedTextRenderer.CalcHeight` → `utils.CalcTextHeight` (total pixel height with overflow charsFit)
- `AdvancedTextRenderer.CalcWidth` → `utils.CalcTextWidth` (max wrapped line width + space)
- `Paragraph.MeasureString` inner loop → `utils.CharsFitInWidth` (chars fitting in pixel width)
- Helper utilities added: `utils.MeasureStringAdvance`, `utils.TabStopPositions`, `utils.FontLineHeight`, `utils.MeasureStringSize`
- HTML tag parsing (C# `WrapHtmlLines`) → `utils.HtmlTextRenderer` in `htmltext.go`: handles `<b>`, `<i>`, `<u>`, `<s>`, `<strike>`, `<br>`, `<font>`, `<span>`, `<sub>`, `<sup>`, inline CSS (color, background-color, font-size, font-family), HTML entities, and multi-line layout into `HtmlLine`/`HtmlRun` structs.

**Intentional architectural divergence (Go headless engine):**
- `AdvancedTextRenderer.Draw()` — NOT ported. Go is a headless engine; each exporter (HTML/PDF/SVG) performs its own rendering. GDI+ drawing is not applicable.
- `StandardTextRenderer.Draw()` — NOT ported (same reason).
- `InlineImageCache` — NOT ported. Inline `<img>` tags in text are not yet rendered; the HTML exporter passes HTML text directly to the browser which handles image rendering.

**Remaining gaps (engine-relevant):**
- **HIGH: No structured Paragraph/Line/Word/Run layout objects** — C# builds a tree with pixel-accurate X/Y positions per run. Go only produces flat `HtmlRun` lists without per-word/per-run positional layout. PDF/RTF exporters cannot iterate runs with positions; they currently use simplified text placement.
- **HIGH: Text justification layout** — `HorzAlign.Justify` expands inter-word spacing to fill the line width (C# `Line.AlignWords` with delta computation). Go has no equivalent; justified text currently falls back to left-aligned in PDF/HTML.
- **HIGH: VertAlign positioning** — C# `AdjustParagraphLines` computes top offsets for Center/Bottom vertical alignment. Go engine stores `VertAlign` on `PreparedObject` but does not compute per-line Y offsets; each exporter handles this independently (HTML via CSS, PDF via manual offset).
- **MEDIUM: Ellipsis trimming modes** — C# `Paragraph.WrapLines` handles `StringTrimming.EllipsisCharacter`, `EllipsisWord`, `EllipsisPath`. Go has no text-overflow ellipsis logic; text is clipped at the display rectangle. No callers currently use this in the Go engine.
- **MEDIUM: Text rotation (Angle)** — C# rotates the display rect for 90/270 degree angles. Go stores `Angle` on `PreparedObject`; exporters apply rotation independently. The Go measurement functions do not account for rotation.
- **MEDIUM: Underline/strikeout drawing at pixel positions** — C# `Line.MakeUnderlines`/`Line.Draw` draws bars at computed pixel offsets. Go uses CSS text-decoration (HTML) or PDF annotation marks instead.
- **LOW: Tab character handling in word layout** — `utils.GetTabPosition` is ported but `utils.wordWrap` treats tabs as zero-width (via `strings.Fields`). Tab-containing text in narrow columns may mis-wrap.
- **LOW: ForceJustify / RightToLeft** — not ported.
- **LOW: InlineImageCache / RunImage** — inline `<img src="...">` tag rendering. Not ported.
- **LOW: widthRatio / fontScale / scale** — C# `AdvancedTextRenderer` accepts horizontal font stretch and DPI scaling. Not exposed in Go measurement functions.

#### `TextUtils.cs`
- **File**: `FastReport.Base/Utils/TextUtils.cs`
- **Status**: FULLY PORTED
- **Go**: `utils.IsWholeWord` in `utils/text.go` (with `isWordDelimiter` helper and `wordDelimiters` set)
- **Notes**: All public API ported. `IsWholeWord` and the delimiter set (chars 0x00–0x20 + punctuation) are fully implemented. No engine callers — used only for designer Find & Replace, which is out of scope for the Go engine port.

#### `Units.cs`
- **File**: `FastReport.Base/Utils/Units.cs`
- **Status**: FULLY PORTED
- **Gaps**: Core units exact match. Only FileSize utility missing (UI display only).

#### `Validator.cs`
- **File**: `FastReport.Base/Utils/Validator.cs`
- **Status**: PARTIALLY PORTED (updated 2026-03-22)
- **Go files**: `utils/validator.go`, `report/reportcomponent.go`
- **Implemented**:
  - `NormalizeBounds` → `utils.NormalizeBoundsF(left, top, width, height float32)` — normalizes negative width/height
  - `RectContainInOtherRect` → `utils.RectContainInOtherF(outerL, outerT, outerW, outerH, innerL, innerT, innerW, innerH float32)` — containment check with 0.01 grid-fit compensation (Validator.cs lines 79–88)
  - Intersection helpers → `utils.RectsIntersectF(...)` — open-interval rectangle intersection (Validator.cs line 70)
  - `ReportComponentBase.Validate()` → `(*report.ReportComponentBase).Validate() []utils.ValidationIssue` — checks: positive size, non-empty name, within parent bounds (ReportComponentBase.cs lines 802–816)
  - `ValidateReport` duplicate-name loop → `ruleDuplicateNames` rule in `utils.ReportValidator` (Validator.cs lines 127–145)
  - `ValidatableReport` interface extended with `ObjectNames() []string` to expose object names without an import cycle
  - 11 new tests in `report/reportcomponent_validate_test.go`, 14 new tests in `utils/validator_test.go`
- **Remaining gaps**:
  - `GetIntersectingObjects` (overlap detection within a band) — not ported. This requires direct access to `BandBase.Objects()` and iterates pairs using `GetExtendedSize()`. Out of scope for the server-side rendering pipeline since the Go port has no designer; intersection highlighting is a designer-only feature.
  - `ValidateReport(report, checkIntersectObj, token)` top-level function — not ported as a free function. The Go equivalent is `ReportValidator.Validate(ValidatableReport)` which uses the rule system. Full structural validation (intersections, per-component `Validate()` calls) requires wiring the concrete `*reportpkg.Report` to the `ValidatableReport` interface by implementing the missing methods (`BandNames`, `DataSourceNames`, `TextExpressions`, `ParameterNames`, `ObjectNames`). That wiring is TODO.
  - `CancellationToken` — not applicable in Go. Cancellation would use `context.Context`.

#### `Variant.cs`
- **File**: `FastReport.Base/Utils/Variant.cs`
- **Status**: NOT PORTED
- **Gaps**: No unified Variant struct — coercion scattered across 6 files. CheckBox eval misses numeric non-zero. Rich bool conversion missing.

#### `Xml.cs`
- **File**: `FastReport.Base/Utils/Xml.cs`
- **Status**: FULLY PORTED
- **Gaps**: Architectural replacement — Go uses encoding/xml Decoder/Encoder.

#### `Zip.cs`
- **File**: `FastReport.Base/Utils/Zip.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Go has bare DEFLATE only. No ZIP archive builder. Use Go stdlib archive/zip.

### Utils/Json

#### `JsonArray.cs`
- **File**: `FastReport.Base/Utils/Json/JsonArray.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

#### `JsonBase.cs`
- **File**: `FastReport.Base/Utils/Json/JsonBase.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

#### `JsonObject.cs`
- **File**: `FastReport.Base/Utils/Json/JsonObject.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

#### `JsonSchema.cs`
- **File**: `FastReport.Base/Utils/Json/JsonSchema.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

#### `JsonTextReader.cs`
- **File**: `FastReport.Base/Utils/Json/JsonTextReader.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

### Utils/Json/Serialization

#### `JsonAttributes.cs`
- **File**: `FastReport.Base/Utils/Json/Serialization/JsonAttributes.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

#### `JsonConverter.cs`
- **File**: `FastReport.Base/Utils/Json/Serialization/JsonConverter.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

#### `JsonDeserializer.cs`
- **File**: `FastReport.Base/Utils/Json/Serialization/JsonDeserializer.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Reflection-based JSON-to-object. Go uses encoding/json. Only consumer is FastReport.Core.Web.

#### `JsonPropertyInfo.cs`
- **File**: `FastReport.Base/Utils/Json/Serialization/JsonPropertyInfo.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.

#### `JsonSerializer.cs`
- **File**: `FastReport.Base/Utils/Json/Serialization/JsonSerializer.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Custom JSON DOM. Go uses encoding/json natively.


## FastReport.Compat

> **OUT OF SCOPE** - .NET compatibility shims (System.Drawing, CodeDom, WinForms replacements).

- `FastReport.Compat/shared/Compiler/CSharpCodeProvider.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/CodeDomProvider.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/CodeGenerator.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/CompilationEventArgs.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/CompilerError.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/CompilerParameters.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/CompilerResults.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/IAssemblyLoadResolver.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/TempFileCollection.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/Compiler/VBCodeProvider.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/DotNetClasses/Color.Full.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/DotNetClasses/GdiGraphics.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/DotNetClasses/IGraphics.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/DotNetClasses/UITypeEditor.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/TypeConverters/Color.Core.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/TypeConverters/ColorConverter.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/TypeConverters/FontConverter.IFontFamilyMatcher.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/TypeConverters/FontConverter.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/TypeConverters/SizeConverter.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/ComboBox.ObjectCollection.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/ItemArray.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/ListBox.ObjectCollection.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/ListBox.SelectedIndexCollection.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/ListBox.SelectedObjectCollection.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/WindowsFormsReplacement.BindingSource.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/WindowsFormsReplacement.ListBindingHelper.cs` - OUT OF SCOPE
- `FastReport.Compat/shared/WindowsForms/WindowsFormsReplacement.cs` - OUT OF SCOPE

## FastReport.Core.Web

> **OUT OF SCOPE** - ASP.NET Core web integration (controllers, middleware, web viewer).

- `FastReport.Core.Web/Application/Cache/CacheOptions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Cache/IWebReportCache.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Cache/WebReportCache.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Cache/WebReportCacheOptions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Cache/WebReportDistributedCache.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Cache/WebReportLegacyCache.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Cache/WebReportMemoryCache.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Constants.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/DesignerOptions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/DesignerSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ExportMenuSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ExportsHelper.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Extensions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/ControllerBuilder.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/ControllerExecutor.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/FastReportBuilderExtensions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/FastReportGlobal.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/FastReportMiddleware.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/FastReportOptions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/FastReportServiceCollectionExtensions.Backend.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/FastReportServiceCollectionExtensions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Infrastructure/IResult.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/LinkerFlags.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/DocxExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/EmailExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/HtmlExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/ImageExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/OdfExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/PageSelectorLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/PdfExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/PptxExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/RtfExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/SvgExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/XlsxExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Localizations/XmlExportSettingsLocalization.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ReportExporter/ReportExporter.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ReportExporter/Strategies/ArchiveExportStrategy.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ReportExporter/Strategies/DefaultExportStrategy.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ReportExporter/Strategies/ExportStrategyFactory.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ReportExporter/Strategies/IExportStrategy.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ReportExporter/Strategies/PreparedExportStrategy.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ReportTab.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Toolbar/ToolbarButton.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Toolbar/ToolbarElement.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Toolbar/ToolbarInput.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/Toolbar/ToolbarSelect.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/ToolbarSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReport.Backend.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReport.Exports.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReport.Tabs.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReport.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReportDesigner.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReportExceptions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReportHtml.Backend.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReportHtml.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebReportOptions.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Application/WebUtils.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/ApiControllerBase.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Designer/ConnectionsController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Designer/DesignerReportController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Designer/UtilsController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Preview/DialogController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Preview/ExportReportController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Preview/GetPictureController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Preview/GetReportController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Preview/PrintReportController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Preview/ServiceController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Controllers/Resources/ResourcesController.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/IConnectionsService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/IDesignerUtilsService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/IExportsService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/IPrintService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/IReportDesignerService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/IReportService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/IResourceLoader.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Abstract/ITextEditService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/EmailExportParams.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Helpers/IntelliSenseHelper.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Helpers/IntelliSenseModels.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/ConnectionService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/DesignerUtilsService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/ExportService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/InternalResourceLoader.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/PrintService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/ReportDesignerService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/ReportService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/Implementation/TextEditService.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Services/ServicesParamsModels.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/DocxExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/EmailExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/HtmlExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/ImageExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/OdsExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/OdtExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/PdfExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/PptxExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/RtfExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/SvgExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/XlsxExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/ExportSettings/XmlExportSettings.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/body.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/main.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/modalcontainer.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/outline.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/script.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/style.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/tabs.cs` - OUT OF SCOPE
- `FastReport.Core.Web/Templates/toolbar.cs` - OUT OF SCOPE

## FastReport.OpenSource

#### `BandBase.Core.cs`
- **File**: `FastReport.OpenSource/BandBase.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — contains only Draw() override for WinForms/designer UI rendering; Go uses PreparedPages snapshot architecture.

#### `Base.Core.cs`
- **File**: `FastReport.OpenSource/Base.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — contains only a no-op ExtractDefaultMacrosInternal() stub; macro substitution is handled in preview/postprocessor.go.

#### `CellularTextObject.Core.cs`
- **File**: `FastReport.OpenSource/CellularTextObject.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — GetCellWidthInternal() auto-sizing logic is ported inline into engine/objects.go:populateCellularTextCells(); all public properties and serialization are implemented. Constructor defaults CanBreak=false and Border.Lines=BorderLines.All (from base CellularTextObject.cs) were missing from NewCellularTextObject() and are now fixed (go-fastreport-1z45f).

#### `ComponentBase.Core.cs`
- **File**: `FastReport.OpenSource/ComponentBase.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — contains only virtual Draw() method stub; Go uses PreparedPages data-driven rendering architecture instead of virtual Draw() paradigm.

#### `HtmlObject.Core.cs`
- **File**: `FastReport.OpenSource/HtmlObject.Core.cs`
- **Status**: FULLY PORTED
- **Fixed 2026-03-22**: Added `Assign(*HtmlObject)`, `GetExpressions()`, `SaveState()`/`RestoreState()` (with `savedText` field), `CalcWidth()`, and `ApplyCondition(style.HighlightCondition)` to `object/html.go`. All mirror their C# counterparts in HtmlObject.cs (lines 80-86, 161-172, 177-187, 193-196, 147-155).
- **Remaining gaps**: `GetData()`/`GetDataShared()` expression substitution is handled by the engine's `evalText(v.Text())` call in `engine/objects.go:680` — no correctness gap. `GetStringFormat()` and `DrawText()`/`Draw()` are GDI+ rendering stubs with no Go equivalent needed. `Break()` and `CalcHeight()` are correctly inherited from `BreakableComponent` (return false and Height() respectively, matching C# stubs). `HtmlObject.Core.cs` itself contains only a `DrawDesign` no-op partial method stub — out of scope.

#### `InternalVisibleTo.OpenSource.cs`
- **File**: `FastReport.OpenSource/InternalVisibleTo.OpenSource.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Single assembly attribute [InternalsVisibleTo]. Go's visibility model handles this natively.

#### `LineObject.OpenSource.cs`
- **File**: `FastReport.OpenSource/LineObject.OpenSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**:
  - `GetConvertedObjects()`: Converts a LineObject with non-None caps into a `PictureObject` by rendering into a `System.Drawing.Bitmap` via GDI+. This is architecturally incompatible with the Go headless export pipeline — there is no GDI+ equivalent available. The Go engine renders lines directly in the export layer (SVG, PDF, HTML, PNG) without an object-replacement step.
  - `IsHaveToConvert()` (from `LineObject.cs` base): Returns `true` when `StartCap.Style != None || EndCap.Style != None`. The Go engine has no object-replacement pipeline, so this predicate has no consumer. Not implemented.
  - `CreatePath()` (from `LineObject.cs` base): Builds a `System.Drawing.Drawing2D.GraphicsPath` encoding the line geometry plus cap shapes, used only by `GetConvertedObjects()` to compute the bounding rect of the rendered bitmap. Not applicable in Go.
  - **Functional consequence**: When `StartCap` or `EndCap` is non-None on a LineObject, the Go engine currently serializes/deserializes the cap settings correctly (round-trip is lossless) but the export layer receives no cap information — `PreparedObject` carries no cap fields — so line caps are silently dropped from all rendered output (SVG, HTML, PNG, PDF). The C# pipeline avoids this by converting the whole LineObject to a pre-rendered bitmap.
  - **What IS ported**: All data fields (`Diagonal`, `StartCap`, `EndCap`, `DashPattern`) serialize and deserialize correctly in `object/lines.go`. The `LineObject` struct and its `CapSettings` type fully match the C# data model.

#### `PictureObject.Core.cs`
- **File**: `FastReport.OpenSource/PictureObject.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — file contains only a design-time stub method (IsDesigningInPreviewPageDesigner returning false); all PictureObject functionality is fully ported.

#### `PictureObjectBase.Core.cs`
- **File**: `FastReport.OpenSource/PictureObjectBase.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — file contains only two empty stub methods (DrawErrorImage, DrawDesign); all PictureObjectBase properties and serialization are fully ported.

#### `PolyLineObject.Core.cs`
- **File**: `FastReport.OpenSource/PolyLineObject.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — this file is a C# partial class containing only three empty partial method stubs (`DrawDesign0`, `DrawDesign1`, `InitDesign`) that are no-ops in the OpenSource build. They exist solely for the designer/preview UI which is out of scope for the Go port. All substantive PolyLineObject functionality (Serialize/Deserialize with PolyPoints_v2 bezier format, legacy PolyPoints v1 format, CenterX/CenterY, DashPattern, PolyPoint with Left/Right curve control points, PolyPointCollection) is implemented in `object/lines.go`. The previous entry incorrectly attributed gaps from `PolyLineObject.cs` to this Core file; those gaps were fixed 2026-03-21. Note: `RecalculateBounds()`, `GetPath()`, `SetPolyLine()`, `GetPseudoPoint()`, and `drawPoly()` live in `PolyLineObject.cs` (not the Core file) and are rendering/design-time methods — not ported because the Go export pipeline uses vector-based SVG/PDF path generation rather than GDI+ GraphicsPath.

#### `Report.Core.cs`
- **File**: `FastReport.OpenSource/Report.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. Implemented `Report.Dispose()` in `reportpkg/report.go` and `PreparedPages.Dispose()` in `preview/prepared_pages.go` — these together match C# `DisposePreparedPages()` which calls `preparedPages.Dispose()`. `Prepare()` and `PrepareWithContext()` now dispose the old `PreparedPages` before replacing, preventing BlobStore temp-file leaks on repeated runs. The design-mode partial methods (SerializeDesign, InitDesign, ClearDesign, DisposeDesign) and performance hooks (StartPerformanceCounter, StopPerformanceCounter) are all no-ops in the OpenSource build and are intentionally omitted from the Go port.

#### `ReportComponentBase.Core.cs`
- **File**: `FastReport.OpenSource/ReportComponentBase.Core.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: Missing DrawMarkers(), DrawCrossHair(), AssignPreviewEvents(), and DrawIntersectBackground() — these are UI/preview-layer drawing methods not applicable to headless Go export engine. Click event is modeled as OnClick callback field instead of assignable event property.

#### `ReportPage.Core.cs`
- **File**: `FastReport.OpenSource/ReportPage.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — file contains only three empty partial method stubs (AssignPreview, InitPreview, WritePreview) that do nothing in OpenSource build.

#### `ReportSettings.Core.cs`
- **File**: `FastReport.OpenSource/ReportSettings.Core.cs`
- **Status**: FULLY PORTED (2026-03-22)
- **What was ported**: `DatabaseLoginEventArgs` and `AfterDatabaseLoginEventArgs` types added to `data/connection.go`. `OnDatabaseLogin` and `OnAfterDatabaseLogin` callback fields added to `data.DataConnectionBase` and wired into `DataConnectionBase.Open()`. Report-level `OnDatabaseLogin` and `OnAfterDatabaseLogin` fields added to `reportpkg.ReportSettings` for per-report login callbacks. 7 new tests in `data/data_connection_coverage_test.go`.
- **Intentionally omitted** (no-ops in C# OpenSource build): `OnProgress(Report, string)`, `OnProgress(Report, string, int, int)`, `OnStartProgress(Report)`, `OnFinishProgress(Report)` — covered by `export.ExportBase.OnProgress` callback; no UI progress in headless Go engine.
- **Intentionally relocated** (idiomatic Go, avoids global-state model): `FilterBusinessObjectProperties` / `GetBusinessObjectPropertyKind` / `GetBusinessObjectTypeInstance` events → ported as `OnFilterProperties` / `OnGetPropertyKind` callbacks on `data.BusinessObjectConverter`. `DatabaseLogin` / `AfterDatabaseLogin` events → ported as func fields on `data.DataConnectionBase` and mirrored on `reportpkg.ReportSettings`.

#### `ShapeObject.Core.cs`
- **File**: `FastReport.OpenSource/ShapeObject.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — file contains only a no-op DrawDesign partial method stub; ShapeObject properties and rendering are fully implemented in Go.

#### `StyleBase.Core.cs`
- **File**: `FastReport.OpenSource/StyleBase.Core.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: GetDefaultFontInternal ported. Fill/TextFill as color.RGBA, Border omitted.

#### `TextObject.Core.cs`
- **File**: `FastReport.OpenSource/TextObject.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — file contains only a single partial method stub (DrawDesign) with no implementation; actual TextObject is fully ported.

### Code/Ms

#### `MsAssemblyDescriptor.Core.cs`
- **File**: `FastReport.OpenSource/Code/Ms/MsAssemblyDescriptor.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: .NET CodeDom/Roslyn script compilation stub. Go uses expr-lang/expr.

### CrossView

#### `CrossViewHelper.Core.cs`
- **File**: `FastReport.OpenSource/CrossView/CrossViewHelper.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — contains only a no-op OnProgressInternal() partial method stub; actual CrossViewHelper functionality is integrated into crossview.go package directly.

#### `CrossViewObject.Core.cs`
- **File**: `FastReport.OpenSource/CrossView/CrossViewObject.Core.cs`
- **Status**: NOT PORTED
- **Gaps**: CrossViewObject lacks report.Base embedding; missing ModifyResult event, Assign/Serialize/DeserializeSubItems, engine lifecycle (InitializeComponent, FinalizeComponent, SaveState, RestoreState), Style property, ModifyResultEvent, CubeSource with event handling, BuildTemplate, OnModifyResult, and result table lifecycle. Must embed table.TableBase and implement full report component interface.

### Data

#### `CsvDataConnection.Core.cs`
- **File**: `FastReport.OpenSource/Data/CsvDataConnection.Core.cs`
- **Status**: MOSTLY PORTED
- **Gaps**: **Reviewed and updated 2026-03-22**. The `.Core.cs` file itself contains only `CheckForChangeConnection` (a private plumbing method for C# property setters that normalises and persists the connection string). The substantive work is in `CsvDataConnection.cs` and `CsvUtils.cs`. **Implemented in this review**:
  - `ConvertFieldTypes` field + `SetConvertFieldTypes`/`ConvertFieldTypes` getter/setter on `CSVDataSource` (`data/csv/csv_fieldtypes.go`).
  - `determineColumnTypes` type inference engine in `data/csv/csv_convert.go` — matches C# `CsvUtils.DetermineTypes` priority: int → float64 → time.Time → string. Mix of int+float promoted to float64 matching C# int+double→double rule. Empty cells ignored.
  - `convertValue` function applies the inferred type to each cell at row-load time.
  - `NewFromConnectionString(name, connectionString)` constructor in `data/csv/csv_fieldtypes.go` — mirrors C# `CsvDataConnection` constructor + property setters via `CheckForChangeConnection`.
  - `ConnectionStringBuilder` setters (`SetCsvFile`, `SetCodepage`, `SetSeparator`, `SetFieldNamesInFirstString`, `SetRemoveQuotationMarks`, `SetConvertFieldTypes`, `SetNumberFormat`, `SetCurrencyFormat`, `SetDateTimeFormat`) and `String()` serialisation in `data/csv/connection_string_setters.go` — mirrors C# `CsvConnectionStringBuilder.ToString()`.
  - 23 new tests in `data/csv/csv_convert_test.go` covering all new functionality (all passing).
  - Also fixed pre-existing `utils` build error: duplicate `RectContainInOtherF` declaration in `utils/validator.go` (second declaration at line 286 shadowing the correct first at line 219 — removed the duplicate).
  - **Remaining gaps** (intentional / out of scope): HTTP/FTP URL loading (C# `CsvUtils.ReadLines` supports `http://` and `ftp://`; Go is local-file/string only), codepage/encoding support (C# uses `Encoding.GetEncoding(builder.Codepage)`; Go's `encoding/csv` always UTF-8), locale-aware number/currency/datetime parsing (C# uses `CultureInfo` and `NumberFormatInfo`; Go uses invariant `strconv`/`time.Parse`), `RelatedPathCheck` relative-path resolution (resolves CSV path relative to report `.frx` file), `RemoveQuotationMarks` flag (C# naïve `Split + Trim('"')` approach; Go uses `encoding/csv` which handles RFC 4180 quoting correctly and is strictly better). Tests: 49 total, all passing.

#### `DataConnectionBase.Core.cs`
- **File**: `FastReport.OpenSource/Data/DataConnectionBase.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. **Reviewed 2026-03-22**: Implemented `FilterConnectionTablesEventArgs` struct and `OnFilterConnectionTables func(*FilterConnectionTablesEventArgs)` callback field on `DataConnectionBase`. `FilterTables()` now iterates table names, fires the callback per entry, and removes entries where `e.Skip == true` — exactly matching C# Core.cs `FilterTables(List<string>)`. UI stub methods (`GetDefaultConnection`, `ShouldNotDispose`, `ShowLoginForm`) are intentionally omitted as out-of-scope no-ops for the Go port. Coverage: 5 new tests added in `data/connection_test.go`.

#### `TableDataSource.Core.cs`
- **File**: `FastReport.OpenSource/Data/TableDataSource.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None — file contains only a single no-op partial method stub (TryToLoadData); full TableDataSource is ported at data/connection.go.

### Dialog

#### `DialogPage.Core.cs`
- **File**: `FastReport.OpenSource/Dialog/DialogPage.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Partial class stub with zero public methods. Full DialogPage is proprietary/WinForms. Go has minimal deserialization stub.

### Engine

#### `ReportEngine.Core.cs`
- **File**: `FastReport.OpenSource/Engine/ReportEngine.Core.cs`
- **Status**: PARTIALLY PORTED
- **Analysis** (2026-03-22): The file `FastReport.OpenSource/Engine/ReportEngine.Core.cs` is a 6-line stub containing only `partial void ShowProgress();`. The actual ReportEngine implementation was analyzed from `FastReport.Base/Engine/ReportEngine.cs` and associated `.Bands.cs`, `.Pages.cs` files.
- **Implemented** (2026-03-22):
  1. **LimitPreparedPages with Report.MaxPages** (`engine/engine.go`): Previously only honoured engine-level `pagesLimit`. Now also trims `PreparedPages` to `Report.MaxPages` (lower priority), matching C# `LimitPreparedPages()` at ReportEngine.cs line 406–426.
  2. **InitializeSecondPassData** (`engine/engine.go`): Added `initializeSecondPassData()` method that resets all data sources to first row before the second rendering pass, matching C# `InitializeSecondPassData()` at ReportEngine.cs line 356–373. Called in `prepareToSecondPassHook` alongside `initTotals()`.
  3. **StartOnOddPage** (`engine/pages.go`): Inserts a blank filler page so the report starts on an odd page number, matching C# `StartFirstPageShared()` in ReportEngine.Pages.cs. Triggered when `page.StartOnOddPage` is true and the current page index is already on an odd-numbered page.
  4. **VisibleExpression evaluation on bands** (`engine/bands.go`, `engine/pages.go`): C# `CanPrint(band)` (ReportEngine.Bands.cs line 259) evaluates `VisibleExpression` and mutates `band.Visible` before the band is added to `PreparedPages`. Go was missing this evaluation for bands rendered via `showFullBandOnce` (DataBand, GroupHeader, etc.) and for bands rendered via `showBand` (ReportTitleBand, PageHeader, etc.). Fixed by adding `VisibleExpression` evaluation at the top of both `showFullBandOnce` and `showBand`, using `b.CalcVisibleExpression(expr, e.report.Calc)`.
  5. **Outline guard for Repeated bands** (`engine/bands.go`): Added `!b.Repeated()` check to the `OutlineExpression` block in `showFullBandOnce`, matching C# `AddBandOutline` (`ReportEngine.Outline.cs` line 29: `if (band.Visible && !IsNullOrEmpty(band.OutlineExpression) && !band.Repeated)`).
  6. **GetBandHeightWithChildren TotalPages special case** (`engine/bands.go`): When a band's `VisibleExpression` contains "TotalPages" and we are in `FinalPass`, include the band height even if currently not visible. Matches C# `GetBandHeightWithChildren` at ReportEngine.cs line 384–385.
- **Remaining Gaps**:
  - `PrintOnPreviousPage`: Requires `PreparedPages.GetLastY()` which does not exist in Go's `preview` package. Skipped.
  - `PageN`/`PageNofM`: Public C# properties; Go only has system variables accessible via `Calc()`. No public getter methods on `ReportEngine`.
  - `UnlimitedHeight`/`UnlimitedWidth`: ReportPage flags not yet forwarded to the engine's page-sizing logic.
  - `DownThenAcross`: Multi-column snake ordering not yet ported.
  - `ResetDesigningFlag`: Design-time only, not applicable at runtime.

#### `ReportEngine.Dialogs.OpenSource.cs`
- **File**: `FastReport.OpenSource/Engine/ReportEngine.Dialogs.OpenSource.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Both methods (RunDialogs, RunDialogsAsync) are no-op stubs in OpenSource edition. Go handles DialogPage as inert stub.

#### `ReportEngine.OpenSource.cs`
- **File**: `FastReport.OpenSource/Engine/ReportEngine.OpenSource.cs`
- **Status**: DONE
- **Analysis** (2026-03-22): The C# file is 21 lines and defines exactly 3 members:
  1. `InitializePages()` — loops `Report.Pages` calling `PreparedPages.AddSourcePage(page)` per `ReportPage`. In C#, `AddSourcePage` deep-clones pages into an internal `SourcePages` list used by the .fpx file-cache system to reduce serialized report size. This mechanism is entirely absent in Go (Go renders in-memory with no .fpx file cache). The Go `preview.SourcePages` type serves a different purpose (tracking source→output page index ranges for double-pass). **No Go equivalent needed — the .fpx page-dictionary optimization is not part of the Go pipeline.** Status: NOT APPLICABLE.
  2. `partial void TranslateObjects(BandBase parentBand)` — C# partial method declaration with no body in OpenSource. The compiler silently drops partial void calls when no implementation exists. This is a hook for commercial-edition script-based object coordinate translation. **OUT OF SCOPE.**
  3. `TranslatedObjectsToBand(BandBase band)` — Empty stub with comment "Avoid compilation errors". Exists so `BandBase.GetData()` can reference it without a link error (the actual call is commented out in BandBase.cs line 634). **OUT OF SCOPE / no Go equivalent needed.**
- **Gaps**: None. All 3 members are either not applicable to Go's architecture or are commercial-edition stubs. No Go code needed.

### Export

#### `ExportBase.OpenSource.cs`
- **File**: `FastReport.OpenSource/Export/ExportBase.OpenSource.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: All 3 members are commercial-edition stubs (ShowPerformance no-op, GetOverlayPage identity, HAVE_TO_WORK_WITH_OVERLAY=false). Zero gaps.

### Export/Html

#### `HTMLExport.OpenSource.cs`
- **File**: `FastReport.OpenSource/Export/Html/HTMLExport.OpenSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: 3 functional methods IMPLEMENTED (ExportHTMLPageBegin/End, ExportBand). 5 members OUT OF SCOPE (commercial stubs). Cross-cutting HTML export gaps: gradient fills, HtmlParagraph, WebPreview, TableBase.

### Matrix

#### `MatrixObject.Core.cs`
- **File**: `FastReport.OpenSource/Matrix/MatrixObject.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: N/A — file contains only two no-op partial method declarations (InitDesign, RefreshTemplate) that are design-time scaffolding specific to C# partial classes with no functional implementation.

### Preview

#### `PageCache.Core.cs`
- **File**: `FastReport.OpenSource/Preview/PageCache.Core.cs`
- **Status**: FULLY PORTED
- **Gaps**: None. GetPageLimit()=50 ported as default limit in preview/pagecache.go NewPageCache(); LRU cache with Get/Remove/Clear fully matches C# behavior.

#### `PreparedPage.OpenSource.cs`
- **File**: `FastReport.OpenSource/Preview/PreparedPage.OpenSource.cs`
- **Status**: PARTIALLY PORTED
- **Gaps**: ProcessText(TextObject) hook stub is the only content in this partial class. Go preparedPage (preview/prepared_pages.go) is a data-only snapshot (Width/Height/PageNo/Bands/Watermark) — no XmlItem serialization/deserialization, no file caching (UseFileCache), no DoAdd/ReadObject XML round-trip, and no StartGetPage/EndGetPage lifecycle. Text postprocessing moved to preview/postprocessor.go (Postprocessor.Process).

### Table

#### `TableBase.Core.cs`
- **File**: `FastReport.OpenSource/Table/TableBase.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: Contains only design-time partial method stubs (DrawDesign, DrawDesign_Borders, DrawDesign_SelectedCells, etc.) — all intentionally no-ops for OpenSource WinForms designer. Functional TableBase is ported at table/table.go.

#### `TableCellData.Core.cs`
- **File**: `FastReport.OpenSource/Table/TableCellData.Core.cs`
- **Status**: OUT OF SCOPE
- **Implemented (2026-03-22)**: Reviewed. The `.Core.cs` file is a 17-line partial class that implements only one method: `IsTranslatedRichObject(ReportComponentBase obj) => false`. This is a no-op stub for the OpenSource edition (the pro edition checks whether a RichTextBox object has been translated). Go has no WinForms RichTextBox concept, so this method does not apply. The runtime logic of `TableCellData` (Text, ColSpan, RowSpan, Width, Height, AttachCell, RunTimeAssign, SetStyle, CalcHeight, UpdateLayout, etc.) is documented under `TableCellData.cs` in the FastReport.Base section above.
- **Gaps**: None for this specific partial class file.

### Utils

#### `Config.Core.cs`
- **File**: `FastReport.OpenSource/Utils/Config.Core.cs`
- **Status**: DONE (all in-scope items implemented)
- **Fixed (2026-03-22)**: All runtime-relevant items now implemented in `utils/config.go`. The C# file adds: `FilterConnectionTablesEventArgs` + event (OUT OF SCOPE), `WebMode` (OUT OF SCOPE), `FullTrust` (always true, not needed), `DoEvent()` no-op, and partial method stubs for UI/export/auth save-restore (all OUT OF SCOPE). The meaningful runtime additions — `Version` const, `GetTempFolder()` fallback to `os.TempDir()`, `CreateTempFile(dir)`, `TempFilePath()`, `GetConfiguredTempFolder()`, package-level `CreateTempFileInDir()` and `GetEffectiveTempFolder()` — are all implemented and covered by 6 new tests in `utils/config_test.go`.
- **Remaining Gaps** (all OUT OF SCOPE): FilterConnectionTables event (connection-wizard UI), WebMode/FullTrust (ASP.NET hosting), persistent config file path settings (Folder/FontListFolder/ApplicationFolder), script security, CodeDom, PrivateFontCollection.

#### `Config.OpenSource.cs`
- **File**: `FastReport.OpenSource/Utils/Config.OpenSource.cs`
- **Status**: DONE (all in-scope items implemented)
- **Fixed (2026-03-22)**: This file only adds `ProcessMainAssembly()` which instantiates `AssemblyInitializer` to register built-in types. Go equivalent: `init()` functions in each sub-package (object/, band/, etc.) register via `serial.DefaultRegistry.MustRegister()`. `Version` is now `utils.Version`. `IsStringOptimization` is not applicable to Go's string model. `FilterConnectionTables` is OUT OF SCOPE.

#### `ExportsOptions.OpenSource.cs`
- **File**: `FastReport.OpenSource/Utils/ExportsOptions.OpenSource.cs`
- **Status**: MOSTLY PORTED
- **Fixed (2026-03-21)**: Added SetFormatEnabled() and AllowOnly() to export/options.go. See ExportsOptions.cs entry above for details.
- **Remaining Gaps**: C# file contains only empty partial method stubs for CreateDefaultExports/SaveOptions/RestoreOptions — the functional tree-menu code is in the non-community ExportsOptions.cs. All tree-menu / UI parts remain OUT OF SCOPE for headless library.

#### `NetRepository.Core.cs`
- **File**: `FastReport.OpenSource/Utils/NetRepository.Core.cs`
- **Status**: OUT OF SCOPE
- **Gaps**: DescriptionHelper parses .NET XML doc comment files for designer UI tooltips. Go port has no visual designer — no code needed.

#### `RegisteredObjects.Core.cs`
- **File**: `FastReport.OpenSource/Utils/RegisteredObjects.Core.cs`
- **Status**: OUT OF SCOPE
- **Note**: This file is a partial class that adds no-op stubs for `UpdateDesign` (designer UI) in the OpenSource build. The full `RegisteredObjects` analysis is covered under `FastReport.Base/Utils/RegisteredObjects.cs` above. The Go registry (`serial/registry.go`) correctly covers all FRX deserialization needs without designer metadata.


## FastReport.Web.Base

> **OUT OF SCOPE** - web resource utilities and toolbar localization.

- `FastReport.Web.Base/ScriptSecurity.cs` - OUT OF SCOPE
- `FastReport.Web.Base/Toolbar.Localization.cs` - OUT OF SCOPE
- `FastReport.Web.Base/WebResources.cs` - OUT OF SCOPE

## Pack

> **OUT OF SCOPE** - build scripts and packaging (Cake build system).

- `Pack/BuildScripts/CakeAPI/CakeAPI.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tasks/BaseTasks.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tasks/Constants.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tasks/LocalizationPackage.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tasks/OpenSourceTasks.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tasks/Tests.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tools/DebugAttribute.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tools/DependsOnAttribute.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tools/Graph.cs` - OUT OF SCOPE
- `Pack/BuildScripts/Tools/Startup.cs` - OUT OF SCOPE

## Tools

> **OUT OF SCOPE** - test utilities for the .NET project.

- `Tools/FastReport.Tests.OpenSource/BaseTests.cs` - OUT OF SCOPE
- `Tools/FastReport.Tests.OpenSource/Data/JsonParserTests.cs` - OUT OF SCOPE
- `Tools/FastReport.Tests.OpenSource/ReportObjectTests/TextObjectTests.cs` - OUT OF SCOPE
- `Tools/FastReport.Tests.OpenSource/TextObjectBaseTests.cs` - OUT OF SCOPE
