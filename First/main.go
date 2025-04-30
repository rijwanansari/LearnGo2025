package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")

	var a, b = 6, "Hello"
	c, d := 7, "World!"

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(d)
	//demoArray()
	demoArray1()
	demoSlice()
}

// This is a simple Go program that prints "Hello, World!" to the console.

//add function for array
func demoArray() {
	var a [5]int
	a[0] = 1
	a[1] = 2
	a[2] = 3
	a[3] = 4
	a[4] = 5

	for i := 0; i < len(a); i++ {
		fmt.Println(a[i])
	}
}

func demoArray1() {
	var arr1 = [5]int{1, 2, 3, 4, 5}
	arr2 := [5]string{"Hello", "World", "!", "Go", "!"}
	fmt.Println(arr1)
	fmt.Println(arr2)
}

// slice
func demoSlice() {
	myslice1 := []int{}
	fmt.Println(len(myslice1))
	fmt.Println(cap(myslice1))
	fmt.Println(myslice1)

	myslice2 := []string{"Go", "Slices", "Are", "Powerful"}
	fmt.Println(len(myslice2))
	fmt.Println(cap(myslice2))
	fmt.Println(myslice2)

}
