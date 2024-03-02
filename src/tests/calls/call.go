package calls

/*
#cgo noescape add2

int add(int a, int b) {
	return a + b;
}

int add2(int a, int b) {
	return a + b;
}
*/
import "C"

func callAdd() int {
	return int(C.add(C.int(1), C.int(2)))
}

func callAdd2() int {
	return int(C.add2(C.int(1), C.int(2)))
}
