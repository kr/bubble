func main() {
	println(&5)
	println((&5)())
	println((&x)(5))
	println((&y)(4, 5))
	println((&z)(3, 4, 5))
}