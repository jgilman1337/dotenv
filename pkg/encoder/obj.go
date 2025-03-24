package encoder

// Represents a set of options for the encoder
type EncoderOpts struct {
	SpacesInArrs  bool //Whether to put spaces after commas in arrays.
	SpaceAroundKV bool //Whether to put spaces around the keys and values.
	IncludePath   bool //Whether to write the path to the element in the resultant dotenv.
	IncludeTyping bool //Whether to write the datatype of the element in the resultant dotenv.
}

// Represents a single dotenv line. Optionally includes info like the struct key and datatype.
type _EnvLine struct {
	Key      string
	Value    string
	Datatype string
	Path     string
}

// Returns the default options for the encoder.
func DefaultOpts() EncoderOpts {
	return EncoderOpts{
		//true, false, false, false,
		true, false, true, false,
	}
}
