/*Package coberturaxml creates a Cobertura XML report from a testjson.Execution.
 */
package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"gotest.tools/assert"
	"gotest.tools/gotestsum/testjson"
)

type CoberturaCoverage struct {
	XMLName         xml.Name `xml:"coverage"`
	LineRate        float32  `xml:"line-rate,attr"`
	BranchRate      float32  `xml:"branch-rate,attr"`
	LinesCovered    int      `xml:"lines-covered,attr"`
	LinesValid      int      `xml:"lines-valid,attr"`
	BranchesCovered int      `xml:"branches-covered,attr"`
	BranchesValid   int      `xml:"branches-valid,attr"`
	Complexity      float32  `xml:"complexity,attr"`
	Version         string   `xml:"version,attr"`
	Timestamp       int64    `xml:"timestamp,attr"`
	Sources         []Source `xml:"sources>source"`
}

type Source struct {
	Filepath string `xml:",chardata"`
}

type Package struct {
	Name       string  `xml:"name,attr"`
	LineRate   float32 `xml:"line-rate,attr"`
	BranchRate float32 `xml:"branch-rate,attr"`
	Complexity float32 `xml:"complexity,attr"`
	Classes    []Class `xml:"classes>class"`
}

type Class struct {
	Name       string    `xml:"name,attr"`
	Filename   string    `xml:"filename,attr"`
	LineRate   float32   `xml:"line-rate,attr"`
	BranchRate float32   `xml:"branch-rate,attr"`
	Complexity float32   `xml:"complexity,attr"`
	Methods    []*Method `xml:"methods>method"`
	Lines      Lines     `xml:"lines>line"`
}

type Method struct {
	Name       string  `xml:"name,attr"`
	Signature  string  `xml:"signature,attr"`
	LineRate   float32 `xml:"line-rate,attr"`
	BranchRate float32 `xml:"branch-rate,attr"`
	Complexity float32 `xml:"complexity,attr"`
	Lines      Lines   `xml:"lines>line"`
}

type Line struct {
	Number int   `xml:"number,attr"`
	Hits   int64 `xml:"hits,attr"`
}

// Lines is a slice of Line pointers, with some convenience methods
type Lines []*Line

type Config struct {
}

func main() {
	c := CoberturaCoverage{}
	doc, err := xml.MarshalIndent(c, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(doc))
}

// Write creates an XML document and writes it to out.
func Write(out io.Writer, exec *testjson.Execution, cfg Config) error {
	if err := write(out, generate(exec, cfg)); err != nil {
		return fmt.Errorf("failed to write JUnit XML: %v", err)
	}
	return nil
}

func write(out io.Writer, cobcov CoberturaCoverage) error {
	doc, err := xml.MarshalIndent(cobcov, "", "\t")
	if err != nil {
		return err
	}
	_, err = out.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	_, err = out.Write(doc)
	return err
}

func generate(exec *testjson.Execution, cfg Config) CoberturaCoverage {
	// cfg = configWithDefaults(cfg)
	// version := goVersion()
	cobcov := CoberturaCoverage{
		XMLName:         xml.Name{},
		LineRate:        0,
		BranchRate:      0,
		LinesCovered:    0,
		LinesValid:      0,
		BranchesCovered: 0,
		BranchesValid:   0,
		Complexity:      0,
		Version:         "",
		Timestamp:       0,
		Sources:         []Source{},
	}
	suites := JUnitTestSuites{
		Name:     cfg.ProjectName,
		Tests:    exec.Total(),
		Failures: len(exec.Failed()),
		Errors:   len(exec.Errors()),
		Time:     formatDurationAsSeconds(time.Since(exec.Started())),
	}

	if cfg.customElapsed != "" {
		suites.Time = cfg.customElapsed
	}
	for _, pkgname := range exec.Packages() {
		pkg := exec.Package(pkgname)
		if cfg.HideEmptyPackages && pkg.IsEmpty() {
			continue
		}
		junitpkg := JUnitTestSuite{
			Name:       cfg.FormatTestSuiteName(pkgname),
			Tests:      pkg.Total,
			Time:       formatDurationAsSeconds(pkg.Elapsed()),
			Properties: packageProperties(version),
			TestCases:  packageTestCases(pkg, cfg.FormatTestCaseClassname),
			Failures:   len(pkg.Failed),
			Timestamp:  cfg.customTimestamp,
		}
		if cfg.customTimestamp == "" {
			junitpkg.Timestamp = exec.Started().Format(time.RFC3339)
		}
		suites.Suites = append(suites.Suites, junitpkg)
	}
	return suites
}

func createExecution(t *testing.T) *testjson.Execution {
	exec, err := testjson.ScanTestOutput(testjson.ScanConfig{
		Stdout: readTestData(t, "out"),
		Stderr: readTestData(t, "err"),
	})
	assert.NilError(t, err)
	return exec
}

func readTestData(t *testing.T, stream string) io.Reader {
	raw, err := ioutil.ReadFile("../../testjson/testdata/input/go-test-json." + stream)
	assert.NilError(t, err)
	return bytes.NewReader(raw)
}
