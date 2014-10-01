func f(k) {
	println("a")
	k(5)
	println("b")
}

func main() {
	println(f(&x))
	println(callcc(f))
}
