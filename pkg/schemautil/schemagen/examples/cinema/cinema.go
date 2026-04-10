package cinema

import (
	"github.com/google/jsonschema-go/jsonschema"
	kubemetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/act3-ai/go-common/pkg/schemautil/schemagen/examples/address"
	"github.com/act3-ai/go-common/pkg/schemautil/schemagen/examples/timestamp"
)

// Cinema represents a cinema.
type Cinema struct {
	// Name of the cinema.
	Name string `json:"name"`
	// Address of the cinema.
	Address *address.Address `json:"address,omitzero"`
	// Theaters in the cinema.
	Theaters []Theater `json:"theaters"`
	// Scheduled showings at the cinema.
	Showings []Showing `json:"showings"`
	// Movies shown at the cinema.
	Movies []Movie `json:"movies"`
}

// Theater represents a theater in a cinema.
type Theater struct {
	// Name of the theater.
	Name string `json:"name"`

	// Number of seats in the theater.
	Seats int `json:"seats"`

	// Layout of the theater.
	Layout []string `json:"layout,omitzero"`
}

type (
	// Showing represents a scheduled showing.
	Showing struct {
		// Name of the theater for the showing.
		Theater string `json:"theater"`

		// Start time of the showing.
		StartTime timestamp.UTCDateTime `json:"startTime"`

		// End time of the showing.
		EndTime timestamp.UTCDateTime `json:"endTime"`

		// Name of the movie being shown.
		Movie string `json:"movie"`
	}

	// Movie represents a movie.
	Movie struct {
		// Type identification.
		kubemetav1.TypeMeta

		// Name of the movie.
		Name string `json:"name"`

		// Runtime of the movie.
		Runtime kubemetav1.Duration `json:"runtime"`

		// Release date of the movie.
		Released timestamp.UTCDate `json:"released"`
	}
)

func (Movie) ExtendJSONSchema(schema *jsonschema.Schema) {
	setConstantTypeMeta(schema, kubemetav1.TypeMeta{
		Kind:       "Movie",
		APIVersion: "cinema/v1",
	})
}

// Rating represents the rating of a show.
type Rating string

//jsonschema:enum
const (
	RatingG       Rating = "G"
	RatingPG      Rating = "PG"
	RatingPG13    Rating = "PG-13"
	RatingR       Rating = "R"
	RatingNC17    Rating = "NC-17"
	RatingUnrated Rating = "UNRATED"
)

// setConstantTypeMeta extends a schema containing a TypeMeta with constant values for the TypeMeta fields.
func setConstantTypeMeta(schema *jsonschema.Schema, typeMeta kubemetav1.TypeMeta) {
	schema.AllOf = append(schema.AllOf, &jsonschema.Schema{
		Const: new(any(typeMeta)),
	})
	// // data, _ := json.MarshalIndent(schema, "", "  ")
	// // fmt.Println(string(data))
	// for _, subschema := range schema.AllOf {
	// 	if strings.HasSuffix(subschema.Ref, "TypeMeta") {
	// 		subschema.Const = new(any(typeMeta))
	// 		return
	// 	}
	// }
	// panic("no TypeMeta found in schema")
}
