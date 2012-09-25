// Copyright © 2012 Steve McCoy under the MIT license.

void test_plus(Testo &t){
	t.Assert(1+1 == 2, "1+1 == 2");
	t.Assert(1+2 == 3, "1+2 == 3");
}

void test_minus(Testo &t){
	t.Assert(1-1 == 0, "1-1 == 0");
	t.Assert(2-1 == 2, "2-1 == 2");
}
