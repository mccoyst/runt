// Copyright © 2012 Steve McCoy under the MIT license.

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

var cxx = flag.String("cxx", "clang++", "The C++ compiler/linker")
var cxxflags = flag.String("cxxflags", "", "Space-separated flags for compilation")
var ldflags = flag.String("ldflags", "", "Space-separated flags for linking")
var testdir = flag.String("testdir", "test", "Location of test files")
var objdir = flag.String("objdir", "build", "Location of object files")
var verbose = flag.Bool("verbose", false, "Print commands to before running them")

var cmdline struct {
	cxx      string
	cxxflags []string
	ldflags  []string
	objects  []string
}

func main() {
	flag.Usage = func(){
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] [additional object files]\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	cmdline.cxx = *cxx
	cmdline.cxxflags = strings.Fields(*cxxflags)
	cmdline.ldflags = strings.Fields(*ldflags)
	cmdline.objects = flag.Args()

	foundObjs, err := filepath.Glob(*objdir + "/*.o")
	if err == nil {
		cmdline.objects = append(cmdline.objects, foundObjs...)
	}

	suites, err := filepath.Glob(*testdir + "/test_*.cpp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find tests: %v\n", err)
		os.Exit(1)
	}

	failures := 0
	for _, suit := range suites {
		if err = runSuite(suit); err != nil {
			fmt.Fprintf(os.Stderr, "Suite failed: %v\n", err)
			failures++
		}
	}
	os.Exit(failures)
}

func runSuite(name string) error {
	code := readTests(name)
	t := &Test{
		string(code),
		findTests(code),
	}

	tmpl := template.Must(template.New("runt").Parse(testfmt))

	testout := name + "_runner.cpp"
	testcpp, err := os.Create(testout)
	if err != nil {
		return fmt.Errorf("Failed to create c++ file: %v.\n", err)
	}
	defer os.Remove(testout)
	defer testcpp.Close()

	err = tmpl.Execute(testcpp, t)
	if err != nil {
		return fmt.Errorf("Failed to write testmain: %v.\n", err)
	}

	fmt.Println("Running", name)

	args := make([]string, 0, len(cmdline.cxxflags)+3+len(cmdline.ldflags))
	args = append(args, cmdline.cxxflags...)
	args = append(args, "-o", "test_runner", testout)
	for _, o := range cmdline.objects {
		args = append(args, o)
	}
	args = append(args, cmdline.ldflags...)

	if *verbose {
		fmt.Println(cmdline.cxx, args)
	}

	cmd := exec.Command(cmdline.cxx, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v. Please fix that.", err)
	}

	if *verbose {
		fmt.Println("./test_runner")
	}

	cmd = exec.Command("./test_runner")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	defer os.Remove("test_runner")
	if err != nil {
		return fmt.Errorf("%v. Please fix that.", err)
	}

	return nil
}

func readTests(file string) []byte {
	testfile, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open %q: %v.\n", file, err)
		os.Exit(1)
	}
	defer testfile.Close()

	code, err := ioutil.ReadAll(testfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read %q: %v.\n", file, err)
		os.Exit(1)
	}
	return code
}

func findTests(code []byte) []string {
	validTest, err := regexp.Compile(`void (test_[[:word:]]+).*`)
	if err != nil {
		panic("Bad regex")
	}

	r := bufio.NewReader(bytes.NewReader(code))
	var tests []string
	for {
		line, isPrefix, err := r.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read code: %v.\n", err)
			os.Exit(1)
		} else if isPrefix {
			fmt.Fprintln(os.Stderr, "Line too long.")
			os.Exit(1)
		}

		if m := validTest.FindSubmatch(line); m != nil {
			tests = append(tests, string(m[1]))
		}
	}
	return tests
}

type Test struct {
	Text  string
	Tests []string
}

var testfmt = `
#include <array>
#include <iostream>
#include <string>
#include <tuple>
#include <utility>
#include <vector>

struct TestFailed{};

class Testo{
	friend int main(int, char*[]);
	std::vector<std::tuple<std::string,int,std::string>> msgs;
	int ntests;
	std::string current;
	Testo() : ntests(0){}
public:
	void Assert(bool b, std::string &&msg){
		ntests++;
		if(b) return;
		msgs.push_back(make_tuple(current, ntests, move(msg)));
		throw TestFailed{};
	}
};

typedef void (test)(Testo &);

{{.Text}}

std::array<std::tuple<std::string, test*>, {{len .Tests}}> tests = { {
{{range .Tests}}
	make_tuple(std::string("{{.}}"), {{.}}),
{{end}}
} };

int main(int argc, char *argv[]){
	Testo testo;
	for(auto t : tests){
		try{
			testo.current = std::get<0>(t);
			std::get<1>(t)(testo);
		}catch(TestFailed tf){
			// OK
		}catch(const std::exception &e){
			std::cout << "Unexpected exception: " << e.what() << '\n';
			return 1;
		}catch(...){
			std::cout << "Unexpected, unknown exception\n";
			return 1;
		}
	}

	if(testo.msgs.empty()){
		std::cout << argv[0] << ": All tests pass.\n";
		return 0;
	}

	std::cout << argv[0] << ": " << testo.msgs.size() << '/' << testo.ntests << " tests failed:\n";
	for(auto msg : testo.msgs)
		std::cout << std::get<0>(msg) << " #"
			<< std::get<1>(msg) << ": \"" << std::get<2>(msg) << "\"\n";
	return 1;
}
`
