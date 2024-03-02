package calls

/*
int add(int a, int b) {
	return a + b;
}
*/
import "C"

func callAdd() int {
	return int(C.add(C.int(1), C.int(2)))
}
