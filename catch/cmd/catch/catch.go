package main

func main() {
	var globalconfig Config

	getenvironmentalvars(&globalconfig)
	printout("CATCH service ready")
	runhttp(&globalconfig)
}
