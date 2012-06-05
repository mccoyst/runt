// Copyright © 2012 Steve McCoy under the MIT license.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"text/template"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "I need a test file.")
		os.Exit(1)
	}

	code := readTests(os.Args[1])
	t := &Test{
		string(code),
		findTests(code),
	}

	tmpl := template.Must(template.New("runt").Parse(testfmt))

	testcpp, err := os.Create(os.Args[1] + ".test.cpp")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create c++ file: %v.\n", err)
		os.Exit(1)
	}
	defer testcpp.Close()

	err = tmpl.Execute(testcpp, t)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write testmain: %v.\n", err)
		os.Exit(1)
	}
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

type Test struct{
	Text string
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
