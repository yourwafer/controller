package life

var initialMethods []func()

func AddAgentInitial(method func()) {
	initialMethods = append(initialMethods, method)
}

func CallInit() {
	for _, method := range initialMethods {
		method()
	}
}
