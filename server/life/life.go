package life

var initialMethods []func()

func AddServerInitial(method func()) {
	initialMethods = append(initialMethods, method)
}

func CallInit() {
	for _, method := range initialMethods {
		method()
	}
}
