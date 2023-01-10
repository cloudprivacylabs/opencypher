package opencypher

import (
	"fmt"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/nleeper/goment"
)

func init() {
	globalFuncs["date"] = Function{
		Name:      "date",
		MinArgs:   0,
		MaxArgs:   1,
		ValueFunc: dateFunc,
	}
	globalFuncs["parseDate"] = Function{
		Name:      "parseDate",
		MinArgs:   2,
		MaxArgs:   2,
		ValueFunc: parseDateFunc,
	}
}

var supportedDateFormats = []string{
	"2006-1-2",
	"200612",
	"2006-1",
	"20061",
	"2006",
}

func dateFunc(ctx *EvalContext, args []Value) (Value, error) {
	if len(args) == 0 {
		return RValue{Value: neo4j.DateOf(time.Now())}, nil
	}
	if len(args) == 1 {
		if args[0].Get() == nil {
			return RValue{}, nil
		}

		if props, ok := args[0].Get().(map[string]Value); ok {
			if value, ok := props["timezone"]; ok {
				str, err := ValueAsString(value)
				if err != nil {
					return nil, err
				}
				loc, err := time.LoadLocation(str)
				if err != nil {
					return nil, err
				}
				return RValue{Value: neo4j.DateOf(time.Now().In(loc))}, nil
			}

			if y, ok := props["year"]; ok {
				if y.Get() == nil {
					return RValue{}, nil
				}
				year, err := ValueAsInt(y)
				if err != nil {
					return nil, err
				}
				day := 1
				if x, ok := props["day"]; ok {
					if x.Get() == nil {
						return RValue{}, nil
					}
					day, err = ValueAsInt(x)
					if err != nil {
						return nil, err
					}
				}
				month := time.January
				if x, ok := props["month"]; ok {
					if x.Get() == nil {
						return RValue{}, nil
					}
					m, err := ValueAsInt(x)
					if err != nil {
						return nil, err
					}
					month = time.Month(m)
				}

				t := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
				return RValue{Value: neo4j.DateOf(t)}, nil
			}
		}

		if str, ok := args[0].Get().(string); ok {
			for _, f := range supportedDateFormats {
				if t, err := time.Parse(f, str); err == nil {
					return RValue{Value: neo4j.DateOf(t)}, nil
				}
			}
			return nil, fmt.Errorf("Invalid date string: %s", str)
		}
	}
	return nil, fmt.Errorf("Invalid use of date function")
}

// parseDate(str,moment format)
func parseDateFunc(ctx *EvalContext, args []Value) (Value, error) {
	if args[1].Get() == nil {
		return RValue{}, nil
	}
	format, err := ValueAsString(args[1])
	if err != nil {
		return nil, fmt.Errorf("In parseDate: %w", err)
	}
	if args[0].Get() == nil {
		return RValue{}, nil
	}
	str, err := ValueAsString(args[0])
	if err != nil {
		return nil, fmt.Errorf("In parseDate:  %w", err)
	}
	g, err := goment.New(str, format)
	if err != nil {
		return nil, fmt.Errorf("In parseDate: %w", err)
	}
	return RValue{Value: neo4j.DateOf(g.ToTime())}, nil
}
