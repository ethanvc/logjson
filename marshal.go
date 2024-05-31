package logjson

func Marshal(in any) []byte {
	return DefaultLogJson().Marshal(in)
}
