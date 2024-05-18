package logjson

func Marshal(in any) []byte {
	return defaultJson.Marshal(in)
}
