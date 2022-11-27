package sqlq

type Operator string

const (
	OperatorEmpty       Operator = "-"
	OperatorAll         Operator = "*"
	OperatorEqual       Operator = "="
	OperatorNotEqual    Operator = "!="
	OperatorGreater     Operator = ">"
	OperatorGreaterThen Operator = ">="
	OperatorLess        Operator = "<"
	OperatorLessThen    Operator = "<="
	OperatorLike        Operator = "LIKE"
	OperatorILike       Operator = "ILIKE"
	OperatorNotLike     Operator = "NOT LIKE"
	OperatorNotILike    Operator = "NOT ILIKE"
	OperatorBetween     Operator = "BETWEEN"
	OperatorIsNull      Operator = "IS NULL"
	OperatorIsNotNull   Operator = "IS NOT NULL"
	OperatorIn          Operator = "IN"
	OperatorNotIn       Operator = "NOT IN"
)

type OperatorsMap map[string][]Operator

func OperatorExists(opr Operator, array []Operator) bool {
	if opr == OperatorEmpty {
		return false
	}

	for _, v := range array {
		if v == OperatorAll || v == opr {
			return true
		}
	}
	return false
}
