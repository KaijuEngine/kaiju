package spec_generator

import "os"

const (
	elmFolder    = "../elements"
	funcFolder   = "../css/functions"
	propFolder   = "../css/properties"
	pseudoFolder = "../css/pseudos"
)

func writeBaseFile(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
