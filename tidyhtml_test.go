package tidyhtml

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

const (
	inSuffix  = ".in.html"
	outSuffix = ".out.html"
)

// A test file has an in.html and out.html version
// and is used for comparing expected output.
type TestFile struct {
	Name, dir string
}

func (tf TestFile) reader(suffix string) io.Reader {
	path := filepath.Join(tf.dir, tf.Name+suffix)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("ReadIn: %s", err)
	}
	// Hack: because my editor adds newlines to the end of files and I
	// would rather not rely on editors doing the desired thing here.
	b = bytes.TrimRightFunc(b, unicode.IsSpace)
	return bytes.NewReader(b)
}

func (tf TestFile) ReadIn() io.Reader {
	return tf.reader(inSuffix)
}

func (tf TestFile) ReadOut() io.Reader {
	return tf.reader(outSuffix)
}

func assertExpected(expected, got io.Reader) error {
	expb, err := ioutil.ReadAll(expected)
	if err != nil {
		panic(err)
	}
	gotb, err := ioutil.ReadAll(got)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(expb, gotb) {
		return stringComparisonError(string(expb), string(gotb))
	}
	return nil
}

func stringComparisonError(expected, got string) error {
	showWhiteSpace := func(s string) string {
		s = strings.Replace(s, " ", ".", -1)
		s = strings.Replace(s, "\t", "....", -1)
		return s
	}
	return fmt.Errorf(
		"Expected:\n[%s]\nGot:\n[%s]",
		showWhiteSpace(expected),
		showWhiteSpace(got),
	)
}

// GetTestFiles finds all ".in.html" test files
// and creates a TestFile for each one.
func GetTestFiles() (testFiles []TestFile) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("os.Getwd: %s", err)
	}
	dir := filepath.Join(cwd, "tests")
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("ReadDir: %s", err)
	}
	for _, fi := range files {
		if strings.HasSuffix(fi.Name(), inSuffix) {
			name := strings.TrimSuffix(fi.Name(), inSuffix)
			testFiles = append(testFiles, TestFile{
				Name: name,
				dir:  dir,
			})
		}
	}
	return
}

func ExampleCopy() {
	r := strings.NewReader(`
		<html>
		<head>
		<title>demo</title>
		</head>
		<body>
		<div><h1>this is a demo</h1>
		<p>
		ok
		</p>
		<div><span>this will do</span></div></div>
		</body></html>
	`)
	if err := Copy(os.Stdout, r); err != nil {
		fmt.Printf("Error: %s", err)
	}
	// Output:
	// <html>
	//     <head>
	//         <title>demo</title>
	//     </head>
	//     <body>
	//         <div>
	//             <h1>this is a demo</h1>
	//             <p>ok</p>
	//             <div>
	//                 <span>this will do</span>
	//             </div>
	//         </div>
	//     </body>
	// </html>
}

func TestCopy(t *testing.T) {
	for _, tf := range GetTestFiles() {

		in := tf.ReadIn()
		out := tf.ReadOut()

		got := bytes.Buffer{}
		w := bufio.NewWriter(&got)
		if err := Copy(w, in); err != nil {
			t.Fatal(err)
		}
		w.Flush()

		if err := assertExpected(out, &got); err != nil {
			t.Errorf("\nFile: %s%s\n%s", tf.Name, outSuffix, err)
		}
	}
}
