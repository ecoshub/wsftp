package json

import (
	"fmt"
	"strings"
    "strconv"
)

func JTM(json string) map[string][]string{
    const (
        QUOTESYBOL  int32 = 34
        OPENBRACE   int32 = 91
        CLOSEBRACE  int32 = 93
        COMMA       int32 = 44
        COL         int32 = 58
        BLANK       int32 = 32
    )
    var currentRune rune;
    json = strings.Trim(json, "{}")
    runes := []rune(json)
    inQuote := false
    inBrace := false

    tempString := ""
    tempInBrace := make([]string, 0, 16)
    nString := 0
    lastKey := ""
    mainMap := make(map[string][]string)
    for i := 0 ; i < len(runes) ; i++ {
        currentRune = runes[i]
        if !inQuote && currentRune == QUOTESYBOL {
            inQuote = true
            continue
        }else if (inQuote && currentRune == QUOTESYBOL) {
            inQuote = false
            if tempString != "" {
                if inBrace {
                    tempInBrace = append(tempInBrace, tempString)
                }else{
                    if nString % 2 == 0 {
                        lastKey = tempString
                    }else{
                        mainMap[lastKey] = []string{tempString}
                    }
                    nString++
                }
                tempString = ""
            }
            continue
        }
        if !inBrace && runes[i] == OPENBRACE {
            inBrace = true
            continue
        }else if inBrace && runes[i] == CLOSEBRACE {
            inBrace = false
            if tempString != ""{
                tempInBrace = append(tempInBrace, tempString)
            }
            mainMap[lastKey] = tempInBrace
            tempInBrace = make([]string, 0, 16)
            tempString = ""
            nString++
        }
        if !inQuote {
            if currentRune != OPENBRACE && currentRune != CLOSEBRACE && currentRune != COMMA && currentRune != COL && currentRune != BLANK{
                tempString += string(currentRune)
            }
            tempString = strings.TrimSpace(tempString)
            if (currentRune == COL || currentRune == COMMA || currentRune == CLOSEBRACE) && tempString != "" {
                if inBrace {
                    tempInBrace = append(tempInBrace, tempString)
                }else{
                    if nString % 2 == 0 {
                        lastKey = tempString
                    }else{
                        mainMap[lastKey] = []string{tempString}
                    }
                    nString++
                }

                tempString = ""
            }
        }else{
            tempString += string(currentRune)
        }
    }
    if tempString != ""{
        mainMap[lastKey] = []string{tempString}
    }
    return mainMap
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