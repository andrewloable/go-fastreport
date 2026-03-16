// FrxToHtml: C# validation tool that uses FastReport .NET (OpenSource) to
// render .frx reports to HTML.  Output is used as ground-truth reference for
// the Go go-fastreport port.
//
// Usage:
//   dotnet run -- --dir ../../test-reports --out ../../csharp-html-output
//   dotnet run -- --frx "../../test-reports/Badges.frx" --out ../../csharp-html-output

using System.Data;
using System.Diagnostics;
using FastReport;
using FastReport.Data;
using FastReport.Export.Html;

// --- Parse arguments --------------------------------------------------------

string frxPath = "";
string frxDir = "../../test-reports";
string nwindPath = "../../test-reports/nwind.xml";
string outDir = "../../csharp-html-output";

for (int i = 0; i < args.Length; i++)
{
    switch (args[i])
    {
        case "--frx" when i + 1 < args.Length:
            frxPath = args[++i];
            break;
        case "--dir" when i + 1 < args.Length:
            frxDir = args[++i];
            break;
        case "--nwind" when i + 1 < args.Length:
            nwindPath = args[++i];
            break;
        case "--out" when i + 1 < args.Length:
            outDir = args[++i];
            break;
    }
}

// --- Load NorthWind DataSet -------------------------------------------------

DataSet? nwindDs = null;
try
{
    nwindDs = new DataSet();
    nwindDs.ReadXml(nwindPath);
}
catch (Exception ex)
{
    Console.Error.WriteLine($"warning: cannot read NorthWind data \"{nwindPath}\": {ex.Message}");
    nwindDs = null;
}

// --- Collect FRX files ------------------------------------------------------

string[] frxFiles;
if (!string.IsNullOrEmpty(frxPath))
{
    frxFiles = [frxPath];
}
else
{
    frxFiles = Directory.GetFiles(frxDir, "*.frx");
    Array.Sort(frxFiles, StringComparer.OrdinalIgnoreCase);
    if (frxFiles.Length == 0)
    {
        Console.Error.WriteLine($"no .frx files found in \"{frxDir}\"");
        return 1;
    }
}

// --- Create output directory ------------------------------------------------

Directory.CreateDirectory(outDir);

// --- Process each report ----------------------------------------------------

int ok = 0, failed = 0;

foreach (var frx in frxFiles)
{
    string baseName = Path.GetFileName(frx);
    string stem = Path.GetFileNameWithoutExtension(frx);
    string outFile = Path.Combine(outDir, stem + ".html");

    try
    {
        var sw = Stopwatch.StartNew();

        using var report = new Report();
        report.Load(frx);

        // Register NorthWind data sources.
        // FRX files reference tables as "NorthWind.Employees" etc., so we
        // register the DataSet with the name "NorthWind" to match.
        if (nwindDs != null)
        {
            report.RegisterData(nwindDs, "NorthWind");

            // Enable all data sources so the engine can iterate them.
            foreach (DataSourceBase ds in report.Dictionary.DataSources)
            {
                ds.Enabled = true;
            }
        }

        report.Prepare();

        using var htmlExport = new HTMLExport
        {
            SinglePage = true,
            Navigator = false,
            EmbedPictures = true,
        };

        report.Export(htmlExport, outFile);

        sw.Stop();
        int pageCount = report.PreparedPages?.Count ?? 0;
        Console.WriteLine($"  OK    {baseName,-50} -> {outFile} ({pageCount} page(s), {sw.ElapsedMilliseconds}ms)");
        ok++;
    }
    catch (Exception ex)
    {
        Console.Error.WriteLine($"  FAIL  {baseName}: {ex.Message}");
        failed++;
    }
}

Console.WriteLine($"\n{ok} succeeded, {failed} failed -- HTML files in \"{outDir}\"");
return failed > 0 ? 1 : 0;
