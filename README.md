Brain-dead unit tester for C++.

Here is an example project layout:

	project/
		src/
			duck.cc
			orange.cc
		build/
			duck.o
			orange.o
		test/
			test_duck.cc
			test_orange.cc

Running the tests with runt may be as simple as:

	runt build/*.o

Although currently, if one of the object files contains `main()`, the test runner
will fail to compile. This will be fixed in the future, but in the meantime, organize your
code so that the main function is easily isolated and filtered out from the other
object files.

Other than that bug, if you have a different project layout or other compilation needs,
they can be specified with flags to runt. For more information, run `runt -h`.
