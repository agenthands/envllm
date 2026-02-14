
package runner

// Skeleton only — implement your model adapter and runtime integration here

type Model interface {
    Complete(prompt string) (string, error)
}

type Case struct {
    ID string `json:"id"`
}

func RunCase(c Case, m Model) error {
    // load prompt
    // send to model
    // parse DSL
    // execute runtime
    // score result
    return nil
}
