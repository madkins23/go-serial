package main

import (
	"fmt"
)

type Wrappable interface{}

type Wrapper[T Wrappable] struct {
	TypeName string
	Contents T
}

func dummyReg(_ interface{}) (string, error) {
	return "dummyReg", nil
}

func Wrap[V Wrappable](contents V) (*Wrapper[V], error) {
	if typeName, err := dummyReg(contents); err != nil {
		return nil, fmt.Errorf("get type name for %#v: %w", contents, err)
	} else {
		return &Wrapper[V]{
			TypeName: typeName,
			Contents: contents,
		}, nil
	}
}

const testDataBytes = 23

type TestData struct {
	bytes []byte
}

func toTestData(text string) TestData {
	return TestData{bytes: ([]byte)(text)}
}

func main() {
	fmt.Println("Hello, playground")
	d := new(TestData)
	one(d)
	two(d)
	three(d)
}

func one(d *TestData) {
	w := &Wrapper[TestData]{
		TypeName: "Wrapper 1",
		Contents: *d,
	}
	fmt.Println("Wrapper 1: ", w.TypeName)
}

func two(_ *TestData) {
	w := &Wrapper[TestData]{
		TypeName: "Wrapper 2",
		Contents: toTestData("shit"),
	}
	fmt.Println("Wrapper 2: ", w.TypeName)
}

func three(d *TestData) {
	if w, err := Wrap(*d); err != nil {
		panic(err)
	} else {
		fmt.Println("Wrapper 3: ", w.TypeName)
	}
}
