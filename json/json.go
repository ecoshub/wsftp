package json

import (
	"fmt"
	"strings"
    "strconv"
    "encoding/json"
)

func JTM(js string) map[string][]string{
    // problem '\' 
    runes := []rune(js)
    esc := rune('\\')
    temp := make([]rune, 0, 2 * len(runes))
    for i := 0 ; i < len(runes) ; i ++ {
        if runes[i] == esc {
            temp = append(temp, esc)
            temp = append(temp, esc)
        }else{
            temp = append(temp, runes[i])
        }
    }
    // fixed '\'
    js = string(temp)
    stringMap := make(map[string]string, 1)
    stringArrayMap := make(map[string][]string, 1)
    stringDec := json.NewDecoder(strings.NewReader(js))
    stringArrayDec := json.NewDecoder(strings.NewReader(js))
    stringDec.Decode(&stringMap)
    stringArrayDec.Decode(&stringArrayMap)
    for k,_ := range stringArrayMap{
        if stringArrayMap[k] == nil {
            stringArrayMap[k] = []string{stringMap[k]}
        }
    }
    return stringArrayMap
}

// Map to JSON
func MTJ(mp map[string][]string) string {
    json := ""
    if len(mp) != 0 {
        json = "{"
        for k, arr := range mp {
            json += fmt.Sprintf(`"%v":%v`, k, listToString(arr))
            json += ","
        }
        json = json[:len(json) - 1]
        json += "}"
    }else{
        json = "[]"
    }
    return json
}

// <NOT PUBLIC> string form list
func listToString(list []string) string{
    lenl := len(list)
    temp := ""
    if list == nil {
        return `""`
    }else{
        if lenl == 0 {
            return `""`
        }else if lenl == 1 {
            if list[0] == "" {
                return `""`
            }
            if GetType(list[0]) == "string" {
                return fmt.Sprintf(`"%v"`, list[0])
            }else{
                return fmt.Sprintf(`%v`, list[0])
            }
        }else if lenl > 1 {
            temp = "["
            for _,v := range list {
                if GetType(v) == "string" {
                    temp += fmt.Sprintf(`"%v",`, v)
                } else{
                    temp += fmt.Sprintf(`%v,`, v)
                }
            }
            temp = temp[:len(temp) - 1] + "]"
            return temp
        }
    }
    return `""`
}

func GetType(val string) string{
	if len(val) > 0{
		if IsBool(val){
			return "bool"
		}else if IsInt(val){
			if val[0] == 48 && len(val) > 1{
				return "string"
			}
			return "int"
		}else if IsFloat(val) {
			return "float"
		}
		return "string"
	}
	return ""
}


func IsBool(val string) bool{
    return val == "true" || val =="false"
}

func IsFloat(val string) bool{
    _, err := strconv.ParseFloat(val, 64)
    if err != nil {
        return false
    } 
    return true
}

func IsInt(val string) bool{
    _, err := strconv.ParseInt(val, 10, 32)
    if err != nil {
        return false
    } 
    return true
}