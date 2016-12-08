package packetDecoding

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"errors"
)

//TODO: RE-Write the GenLayers to use the new packet DECODED (all spacing was removed, so its pure hex now)

type LayerInfo struct {
	Start int 	`json:"Start"`
	End int 	`json:"End"`
	Size int	`json:"Size"`
}

//Struct to store the data from Layer2
type Layer2 struct {
	Info      *LayerInfo	`json:Info`
	////Starting byte index of the layer from the hex dump
	//start     int 	`json:"start"`
	////Ending byte index of the layer from the hex dump
	//End       int	`json:"End"`
	////size (# of HEX) of the layer
	//size      int	`json:"size"`
	//Destination type
	DestType  int	`json:"DestType"`
	// Number of 0's to pad out the frame
	PadZeros  int	`json:"PadZeros"`
	// Number of blocks (of 60 bits) per frame
	NumBlocks int	`json:"NumBlocks"`
	// CRC check for layer 2 (we don't use this...), store the hex w/o processing
	Crc       string	`json:"Crc"`
}


//Struct to store the data from Layer3
type Layer3 struct {
	Info       *LayerInfo	`json:Info`
	//
	////Starting byte index of the layer from the hex dump
	//start int	`json:"start"`
	////Ending byte index of the layer from the hex dump
	//end int		`json:"end"`
	////size (# of bytes) of the layer
	//size int	`json:"size"`

	//header
	//vale of Q - should always be 0...
	Q             int	`json:"Q"`
	//value of D - conf packet required
	D             bool	`json:"D"`
	// packet type
	PacketType    string	`json:"PacketType"`
	// priority of the packet
	Priority      int	`json:"Priority"`
	//RF ack disabled or enabled
	RfAck         bool	`json:"RfAck"`
	// channel - should always be 0
	Channel       int `json:"Channel"`
	// Tx or SSeq number
	Tx            int	`json:"Tx"`
	// RX or Rseq
	Rx            int	`json:"Rx"`
	// length of source address
	LenSrc        int	`json:"LenSrc"`
	// length of the destination address
	LenDest       int	`json:"LenDest"`
	// destination address
	DestAddr      string	`json:"DestAddr"`
	// source address
	SourceAddr    string	`json:"SourceAddr"`
	// Fil3 and facility length (usually 0)
	Fil3          int	`json:"Fil3"`
	LenFacility   int	`json:"LenFacility"`

	//end index of the layer
	LayerEndIndex int	`json:"LayerEndIndex"`
}

//Struct to store the data from Layer4
type Layer4to7 struct {
	Info         *LayerInfo	`json:Info`

	////Starting byte index of the layer from the hex dump
	//start int	`json:"Q"`
	////Ending byte index of the layer from the hex dump
	//end int
	////size (# of bytes) of the layer
	//size int
	// message number
	MessNum      int	`json:MessNum`
	// more parts?
	More         bool	`json:More`
	// part number
	PartNum      int	`json:PartNum`
	// End to End ACK required?
	E2eAck       bool	`json:E2EAck`
	// number of parts of message
	NumParts     int	`json:NumParts`
	// message vital?
	Vital        bool	`json:Vital`
	// message label
	Label        string	`json:Label`
	// revision level (3 is current)
	RevLvl       int	`json:RevLvl`
	// number of bytes of data
	NumBytes     int	`json:NumBytes`
	//number of data bits in the last octet
	NumLastBits  int	`json:NumLastBits`
	// Code line data as a string of bytes, for later processing.
	CodeLineData string	`json:CodeLineData`
	// CRC 16 check (we don't use this) store as a string of hex.
	Crc          string	`json:Crc`
}

//Struct to store all of the Layer information
type TrainInfo struct {
	//hexDump string
	L2  Layer2
	L3  Layer3
	L4P Layer4to7
}

func NewTrainInfo(hex string) *TrainInfo {
	t := new(TrainInfo)
	//t.hexDump = hex
	t.L2 = GenLayer2(hex)

	var err error
	t.L3, err = GenLayer3(hex, t.L2.Info.End)
	if err != nil { // Then stop
		return nil
	}
	t.L4P = GenLayer4to7(hex, t.L3.LayerEndIndex)
	return t
}

func (t *TrainInfo) String() string {
	return fmt.Sprintf("Layer2: %v | Layer3: %v | Layer4to7: %v\n", t.L2, t.L3, t.L4P)
}

//Function to generate Layer2 information
// size is always 10 bytes!
func  GenLayer2(hex string) Layer2{
	l2 := Layer2{}
	info := new(LayerInfo)
	//println("Hex Dump: ", hex)
	info.Start = 0

	//Destination Address
	//first 9 chars are just line numbers and spacing, so skip them.
	str := hex[0:2]
	dec := HexToDec(str)
	l2.DestType = dec

	// # 0's
	str = hex[2:4]
	dec = HexToDec(str)
	l2.PadZeros = dec

	// # Blocks
	str = hex[4:6]
	dec = HexToDec(str)
	l2.NumBlocks = dec

	// CRC
	l2.Crc = hex[6:10]

	info.End = 10
	info.Size = 10 // 10 bytes = 2 bytes + 2 + 2 + 4
	l2.Info = info
	return l2
}

func  GenLayer3(hex string, start int) (Layer3, error){
	l3 := Layer3{}
	info := new(LayerInfo)
	info.Start = start
	i := start
	e := start+2
	str := hex[i:e]
	bin := []byte(HexToBinary(str))
	//Q is always 0 and is cut off from the string...
	l3.Q = 0
	//d
	// array indexing is backwards compared to bit[] indexing...
	if string(string(bin[0])) == "1" {
		l3.D = true
	} // false by default...
	//Type
	tip := string(bin[1:3])
	if  tip == "10" {
		l3.PacketType = "Info"
	} else if tip == "00" {
		l3.PacketType = "Nack"
	} else if tip == "11" {
		l3.PacketType = "Ack"
	} else {
		e := fmt.Sprintln("ERROR: Packet type = ", tip)
		return l3, errors.New(e)
	}

	// priority
	pS := string(bin[3:6])
	l3p64, _ := strconv.ParseInt(pS, 2, 32)
	l3.Priority = int(l3p64)
	//RF ack disabled?
	if string(bin[6]) == "0" {
		l3.RfAck = true // not disabled = enabled
	}

	// Channel - should be constant 00
	i = e
	e = i+2
	str = hex[i:e]
	if str != "00" {
		e := fmt.Sprintln("ERROR, channel should be 00!")
		return l3, errors.New(e)
	}
	l3.Channel = 0

	// TX or SSeq Number
	i = e
	e = i+2
	str = hex[i:e]
	bin = []byte(HexToBinary(str))
	//last bit is always 0, only use the first 7 bits for number
	if len(bin) == 8 { //go cuts out any leading 0's, so we need to check the size
		l3p64, _ = strconv.ParseInt(string(bin[:7]), 2, 32)
	} else {
		l3p64, _ = strconv.ParseInt(string(bin), 2, 32)
	}
	l3.Tx = int(l3p64)

	//RX or Rseq number
	i = e
	e = i+2
	str = hex[i:e]
	bin = []byte(HexToBinary(str))
	//last bit is always 0, only use the first 7 bits for number
	if len(bin) == 8 { //go cuts out any leading 0's, so we need to check the size
		l3p64, _ = strconv.ParseInt(string(bin[:7]), 2, 32)
	} else {
		l3p64, _ = strconv.ParseInt(string(bin), 2, 32)
	}
	l3.Rx = int(l3p64)

	// length of source and destination
	i = e
	e = i+2
	str = hex[i:e]
	dec := HexToDec(string(str[0]))
	l3.LenSrc = dec // nibbles = 4 bits * the decimal number
	dec = HexToDec(string(str[1]))
	l3.LenDest = dec // nibbles = 4 bits * the decimal number

	replace := func(str string) string{
		str = strings.Replace(str, " ", "", -1)
		str = strings.Replace(str, "a", "0", -1)
		str = strings.Replace(str, "b", "1", -1)
		str = strings.Replace(str, "c", "2", -1)
		str = strings.Replace(str, "d", "3", -1)
		str = strings.Replace(str, "e", "4", -1)
		str = strings.Replace(str, "f", "5", -1)
		return str
	}

	// destination address
	i = e
	e = i+l3.LenDest
	str = hex[i:e]
	str = replace(str)
	l3.DestAddr = str
	//fmt.Println("Dest Address: " + str)

	//Source address
	i = e
	e = i+l3.LenSrc
	str = hex[i:e]
	str = replace(str)
	l3.SourceAddr = str
	//fmt.Println("Source address: ", str)

	//Fil3
	i = e
	e = i+1
	str = hex[i:e]
	l3.Fil3 = HexToDec(str)

	i = e
	e = i+1
	str = hex[i:e]
	l3.LayerEndIndex = e
	l3.LenFacility = HexToDec(str)
	if str != "0" {
		e := fmt.Sprintln("ERROR WITH FACILITY LEN")
		return l3, errors.New(e)
	}
	info.End = e
	info.Size = info.End - info.Start
	l3.Info = info

	return l3, nil
}

func GenLayer4to7(hex string, start int) Layer4to7{
	l4p := Layer4to7{}
	info := new(LayerInfo)
	info.Start = start
	//l4p.start = start
	currIndex:=start
	endIndex:=start+2
	//Make the first slice -- Message number and whether or not there are more parts
	str := hex[currIndex:endIndex]
	//convert those digits into binary string
	binary := HexToBinary(str)
	//slice the resulting binary string into the message number and more parts section
	msgNumBinStr := binary[0:7]
	morePartsBinStr := binary[7:]
	//check whether more parts binary bit is true or false and set it
	if morePartsBinStr == "1"{
		l4p.More = true
	}else {
		l4p.More = false
	}
	//Convert the message number from binary string to decimal int
	msgNumDecInt, err := strconv.ParseInt(msgNumBinStr, 2, 16)
	if err != nil{
		panic(err)
	}
	//set the message number value and print both values
	l4p.MessNum = int(msgNumDecInt)

	currIndex=endIndex
	endIndex = currIndex+2
	//make second slice -- Part number and END-TO-END ACK
	str = hex[currIndex:endIndex]
	binary = HexToBinary(str)
	//slice the resulting binary string into the part number and END-TO-END ACK
	partNumBinStr := binary[0:7]
	e2eAckBinStr := binary[7:]
	//Check whether the END-TO-END ACK bit is true or false and set it
	if e2eAckBinStr == "1"{
		l4p.E2eAck = true
	}else {
		l4p.E2eAck = false
	}
	//Convert the part number from binary string to decimal int
	partNumDecInt, err := strconv.ParseInt(partNumBinStr, 2, 16)
	if err != nil{
		panic(err)
	}
	//Set the part number and print both values
	l4p.PartNum = int(partNumDecInt)

	currIndex=endIndex
	endIndex = currIndex+2
	//make third slice -- Number of parts and message vitality
	str = hex[currIndex:endIndex]
	//fmt.Println("Hex3: ", str)
	binary = HexToBinary(str)
	//slice the resulting binary string into the number of parts and whether or not it is vitaL
	numPartsBinStr := binary[0:7]
	vitalBinStr := binary[7:]
	//Check the last bit to see if message is vital and set it
	if vitalBinStr == "1"{
		l4p.Vital = true
	}else {
		l4p.Vital = false
	}
	//Convert the number of parts from binary string to decimal int
	numPartsDecInt, err := strconv.ParseInt(numPartsBinStr, 2, 16)
	if err != nil{
		panic(err)
	}
	//Set the number of parts and print both values
	l4p.NumParts = int(numPartsDecInt)

	// label: int 64 of 4 hex.
	currIndex=endIndex
	endIndex = currIndex+4
	str = string(hex[currIndex:endIndex])
	i64, err := strconv.ParseInt(str, 16, 32)
	label64 := "9."
	i32 := int(i64) - 9*512
	t := i32/64
	label64 += strconv.Itoa(t) + "."
	t = i32 % 64
	label64 += strconv.Itoa(t)
	l4p.Label = label64

	currIndex=endIndex
	endIndex = currIndex+2
	//make fifth slice -- rev level
	str = hex[currIndex:endIndex]
	l4p.RevLvl = HexToDec(str)

	//Bits 58 and 59 are skipped
	//NOTE: We skip 2 zero's here!
	currIndex=endIndex +2
	endIndex = currIndex+2

	//make sixth slice -- number of octets in the data
	str = hex[currIndex:endIndex]
	l4p.NumBytes = HexToDec(str)

	currIndex=endIndex
	endIndex = currIndex+2
	//make seventh slice -- number of data bits in the last octet
	str = hex[currIndex:endIndex]
	l4p.NumLastBits = HexToDec(str)

	currIndex=endIndex
	endIndex = currIndex+l4p.NumBytes *2 // 2 octets per hex

	//make sixth slice -- number of octets in the data
	str = hex[currIndex:endIndex]
	//set the value and print it
	l4p.CodeLineData = str

	//make the last slice -- CRC 16 check. Just need to store string version of the hex
	currIndex=endIndex
	endIndex = currIndex+4
	//make seventh slice -- number of data bits in the last octet
	str = hex[currIndex:endIndex]

	l4p.Crc = str
	info.End = endIndex

	l4p.Info = info
	return l4p
}

//func  GenLayer3(hex string) Layer3{
//	l3 := Layer3{}
//	//println("Hex Dump: ", hex)
//
//	//TODO start of layer 3 header (should be 64, or can it change?)
//	str := hex[25:27]  // Maybe convert this to bit array, similar to how he does in his notes.
//	bin := []byte(HexToBinary(str))
//	//println("Hex: ", str)
//	//println("Dec: ", string(bin))
//	//println("Length : ", len(bin))
//	//Q is always 0 and is cut off from the string...
//	l3.Q = 0
//	//d
//	// array indexing is backwards compared to bit[] indexing...
//	if string(string(bin[0])) == "1" {
//		l3.d = true
//	} // false by default...
//	//Type
//	tip := string(bin[1:3])
//	if  tip == "10" {
//		l3.packetType = "Info"
//	} else if tip == "00" {
//		l3.packetType = "Nack"
//	} else if tip == "11" {
//		l3.packetType = "Ack"
//	} else {
//		fmt.Println("ERROR: Packet type = ", tip)
//	}
//
//	// priority
//	pS := string(bin[3:6])
//	l3p64, _ := strconv.ParseInt(pS, 2, 32)
//	l3.priority = int(l3p64)
//	//RF ack disabled?
//	if string(bin[6]) == "0" {
//		l3.rfAck = true // not disabled = enabled
//	}
//
//	// Channel - should be constant 00
//	str = hex[28:30]
//	//println("Hex: ", str)
//	//println("Dec: ", HexToDec(str))
//	if str != "00" {
//		fmt.Println("ERROR, channel should be 00!")
//	}
//	l3.channel = 0
//
//	// TX or SSeq Number
//	str = hex[31:33]
//	//println("Hex: ", str)
//	//println("Dec: ", HexToDec(str))
//	bin = []byte(HexToBinary(str))
//	//last bit is always 0, only use the first 7 bits for number
//	if len(bin) == 8 { //go cuts out any leading 0's, so we need to check the size
//		l3p64, _ = strconv.ParseInt(string(bin[:7]), 2, 32)
//	} else {
//		l3p64, _ = strconv.ParseInt(string(bin), 2, 32)
//	}
//	l3.tx = int(l3p64)
//	//println(hex[31:])
//	//RX or Rseq number
//	str = hex[35:37]
//	//println(str)
//	bin = []byte(HexToBinary(str))
//	//last bit is always 0, only use the first 7 bits for number
//	if len(bin) == 8 { //go cuts out any leading 0's, so we need to check the size
//		l3p64, _ = strconv.ParseInt(string(bin[:7]), 2, 32)
//	} else {
//		l3p64, _ = strconv.ParseInt(string(bin), 2, 32)
//	}
//	l3.rx = int(l3p64)
//
//	// length of source and destination
//	str = hex[38:40]
//	//fmt.Println(str)
//	//fmt.Println(string(str[0]))
//	dec := HexToDec(string(str[0]))
//	//fmt.Println(dec)
//	l3.lenSrc = dec // nibbles = 4 bits * the decimal number
//	dec = HexToDec(string(str[1]))
//	//fmt.Println(dec)
//	l3.lenDest = dec // nibbles = 4 bits * the decimal number
//
//	// destination address
//	//dEnd := 41+l3.lenDest+((l3.lenDest/2)-1) // must account for spaces.
//	//fmt.Printf("DEND: %d, len Dest: %d len src: %d \n", dEnd, l3.lenDest, l3.lenSrc)
//	//str = hex[41:dEnd]
//	//l3.destAddr = str
//	//fmt.Println("Dest Address: " + str)
//	//sStart := (dEnd + 1)
//
//	//i:=sStart + 3
//
//	//begin with a blank string where we will build the destination address
//	str=""
//	//know that the starting index is 41
//	i:=41
//
//	//have two strings representing the current character to look at and the previous character,
//	//to ensure that no extra spaces get input
//	currChar:=""
//	prevChar:=""
//	//loop is entered on the condition that the address is no more than the length it should be plus the
//	//spaces in between
//	for ; len(str) < l3.lenDest+ ((l3.lenDest / 2) - 1); {
//		currChar = string(hex[i])
//		prevChar = string(hex[i-1])
//
//		//if both characters are spaces just move on to the next character
//		if prevChar == " " && currChar == " "{
//			i++
//
//		//if the character at i is not equal to | then append it to the string (| is the start of unintelligible
//			//code that is irrelevant to our deciphering
//		}else if string(hex[i]) != "|" {
//			str += string(hex[i])
//			i++
//
//			//else is entered when | is encountered
//		}else{
//			//i is incremented, and then the loop skips over all characters inside the | |
//			i++
//			for ; string(hex[i]) != "|" ; {
//				i++
//			}
//			//skips i ahead past the line denotation in the dump
//			i += 12
//
//		}
//	}
//	str = strings.Replace(str, " ", "", -1)
//	str = strings.Replace(str, "A", "0", -1)
//	str = strings.Replace(str, "B", "1", -1)
//	str = strings.Replace(str, "C", "2", -1)
//	str = strings.Replace(str, "D", "3", -1)
//	str = strings.Replace(str, "E", "4", -1)
//	str = strings.Replace(str, "F", "5", -1)
//	//the destination address becomes the string that we built
//	l3.destAddr = str
//	fmt.Println("Dest Address: " + str)
//	//str is reset to now build the source address
//	str = ""
//
//	//sourceStartStr := hex[sStart:sStart + 2]
//	//str = sourceStartStr
//	//currChar=""
//	//prevChar=""
//	//fmt.Println("Previous: " + prevChar)
//	//fmt.Println("Current: " + currChar)
//
//	i++
//	fmt.Println("Hex DUmp pre Loop: ", hex[i:])
//	//loop works exactly the same as the destination address loop
//	for ; len(str) < l3.lenSrc + ((l3.lenSrc / 2) - 1); {
//		currChar = string(hex[i])
//		prevChar = string(hex[i-1])
//
//		if prevChar == " " && currChar == " "{
//			i++
//		}else if string(hex[i]) != "|" {
//			str += string(hex[i])
//			i++
//		}else{
//			i++
//			for ; string(hex[i]) != "|" ; {
//				i++
//			}
//			i += 12
//			//fmt.Println("hex at i: " + string(hex[i]))
//		}
//	}
//
//	//sEnd := sStart + l3.lenSrc + ((l3.lenSrc/2)-1)
//	//end of the source address in the hex dump is now at index i
//	sEnd := i
//	//fmt.Println("sEnd: ", sEnd)
//	str = strings.Replace(str, " ", "", -1)
//	str = strings.Replace(str, "A", "0", -1)
//	str = strings.Replace(str, "B", "1", -1)
//	str = strings.Replace(str, "C", "2", -1)
//	str = strings.Replace(str, "D", "3", -1)
//	str = strings.Replace(str, "E", "4", -1)
//	str = strings.Replace(str, "F", "5", -1)
//	//str = hex[sStart:sEnd]
//	//sets the source address to our current string
//	l3.SourceAddr = str
//
//	fmt.Println("Source address: ", str)
//
//	//Fil3
//	str = hex[sEnd + 1: sEnd+2]
//	//fmt.Println("FIL3:", str)
//	l3.fil3 = HexToDec(str)
//	fmt.Println("HEX DUMP:", hex)
//	fmt.Println("Layer 3 INDEX: ", hex[sEnd:])
//	fmt.Println("Str:", str, "indices:", sEnd+1, ", ", sEnd+2)
//
//	//facility length
//	str = hex[sEnd + 2 : sEnd + 3]
//	fmt.Println("Str:", str, "indices:", sEnd+2, ", ", sEnd+3)
//	l3.LayerEndIndex = sEnd+ 4
//	//fmt.Println("FacLen:", str)
//	l3.lenFacility = HexToDec(str)
//	return l3
//}
//
//func GenLayer4to7(hex string, start int) Layer4to7{
//	l4p := Layer4to7{}
//	println("Hex Dump: ", hex)
//
//	//Declare the variables we know already
//	l4p.start = start// 46 - notice that 46 was only for the example, in reality, we want to pick up from the end of layer 3.
//	//l4p.end = 81
//	//l4p.size = 36
//	currIndex:=start
//	endIndex:=start+2
//	//Make the first slice -- Message number and whether or not there are more parts
//	str := hex[currIndex:endIndex]
//	//fmt.Println("Hex1: ", str)
//	//convert those digits into binary string
//	binary := HexToBinary(str)
//	fmt.Println("Binary Length: ",  len(binary))
//	//slice the resulting binary string into the message number and more parts section
//	msgNumBinStr := binary[0:7]
//	morePartsBinStr := binary[7:]
//	//check whether more parts binary bit is true or false and set it
//	if morePartsBinStr == "1"{
//		l4p.more = true
//	}else {
//		l4p.more = false
//	}
//	//Convert the message number from binary string to decimal int
//	msgNumDecInt, err := strconv.ParseInt(msgNumBinStr, 2, 16)
//	if err != nil{
//		panic(err)
//	}
//	//set the message number value and print both values
//	l4p.messNum = int(msgNumDecInt)
//	fmt.Println("Message num: ", msgNumDecInt)
//	fmt.Println("More parts: ", morePartsBinStr)
//
//
//	currIndex=endIndex+2
//	endIndex+=4
//	//make second slice -- Part number and END-TO-END ACK
//	str = hex[currIndex:endIndex]
//	//fmt.Println("Hex2: ", str)
//	binary = HexToBinary(str)
//	//fmt.Println("Hex2 Bin: ", binary )
//	//slice the resulting binary string into the part number and END-TO-END ACK
//	partNumBinStr := binary[0:7]
//	e2eAckBinStr := binary[7:]
//	//Check whether the END-TO-END ACK bit is true or false and set it
//	if e2eAckBinStr == "1"{
//		l4p.e2eAck = true
//	}else {
//		l4p.e2eAck = false
//	}
//	//Convert the part number from binary string to decimal int
//	partNumDecInt, err := strconv.ParseInt(partNumBinStr, 2, 16)
//	if err != nil{
//		panic(err)
//	}
//	//Set the part number and print both values
//	l4p.partNum = int(partNumDecInt)
//	fmt.Println("Part number: ", partNumDecInt)
//	fmt.Println("Ack e2e: ", e2eAckBinStr)
//
//
//	currIndex=endIndex+1
//	endIndex+=3
//	//make third slice -- Number of parts and message vitality
//	str = hex[currIndex:endIndex]
//	//fmt.Println("Hex3: ", str)
//	binary = HexToBinary(str)
//	//slice the resulting binary string into the number of parts and whether or not it is vitaL
//	numPartsBinStr := binary[0:7]
//	vitalBinStr := binary[7:]
//	//Check the last bit to see if message is vital and set it
//	if vitalBinStr == "1"{
//		l4p.vital = true
//	}else {
//		l4p.vital = false
//	}
//	//Convert the number of parts from binary string to decimal int
//	numPartsDecInt, err := strconv.ParseInt(numPartsBinStr, 2, 16)
//	if err != nil{
//		panic(err)
//	}
//	//Set the number of parts and print both values
//	l4p.numParts = int(numPartsDecInt)
//	fmt.Println("Num parts: ", numPartsDecInt)
//	fmt.Println("Vital: ", vitalBinStr)
//
//
//	currIndex=endIndex+1
//	endIndex+=3
//	//make fourth slice -- Label
//	//First convert the hex to decimal, then follow this formula:
//	//Divide the decimal by 512 to get the first part of the label
//	//Divide the remainder of the first divison by 64 to get the second part
//	//Take whatever remainder is left (this is the third part) i.e. 128B = 4747, 4747 = 9 * 512 + 2 * 64 + 11(remainder from those operations), so the label is: "9.2.11"
//	str = string(hex[currIndex:endIndex])
//	currIndex+=3
//	endIndex+=3
//	str+=string(hex[currIndex:endIndex])
//	//fmt.Println("Hex4: ", str)
//	//convert from hex to decimal
//	labelDecInt, err := strconv.ParseInt(str, 16, 16)
//	if err != nil{
//		panic(err)
//	}
//	//Divide the decimal by 512 and save the remainder
//	labelPt1 := labelDecInt / 512
//	firstRem := labelDecInt % 512
//	//Divide the remainder by 64 and save the remainder
//	labelPt2 := firstRem / 64
//	finalRem:= firstRem % 64
//	//Format the label
//	label := fmt.Sprintf("%d.%d.%d", labelPt1, labelPt2, finalRem)
//	//Print and set the value
//	l4p.label = label
//	fmt.Println("Label: ", label)
//
//
//	currIndex=endIndex+1
//	endIndex+=3
//	//make fifth slice -- rev level
//	str = hex[currIndex:endIndex]
//	//fmt.Println("Hex5: ", str)
//	revLvl := HexToDec(str)
//	//set the value and print it
//	l4p.revLvl = revLvl
//	fmt.Println("revLvl: ", revLvl)
//
//	//Bits 58 and 59 are skipped
//	currIndex=endIndex+4
//	endIndex+=6
//	//make sixth slice -- number of octets in the data
//	str = hex[currIndex:endIndex]
//	//fmt.Println("Hex6: ", str)
//	numOctets := HexToDec(str)
//	//set the value and print it
//	l4p.numBytes = numOctets
//	fmt.Println("numBytes: ", numOctets)
//
//
//	currIndex=endIndex+1
//	endIndex+=3
//	//make seventh slice -- number of data bits in the last octet
//	str = hex[currIndex:endIndex]
//	//fmt.Println("Hex7: ", str)
//	numLastBits := HexToDec(str)
//	//set the value and print it
//	l4p.numLastBits = numLastBits
//	fmt.Println("numLastBytes: ", numLastBits)
//
//	currChar:=""
//	prevChar:=""
//	str=""
//	for ; len(str) < 20; {
//		currChar = string(hex[currIndex])
//		prevChar = string(hex[currIndex-1])
//
//		if prevChar == " " && currChar == " "{
//			currIndex++
//		}else if string(hex[currIndex]) != "|" {
//			str += string(hex[currIndex])
//			currIndex++
//		}else{
//			currIndex++
//			for ; string(hex[currIndex]) != "|" ; {
//				currIndex++
//			}
//			currIndex += 12
//			//fmt.Println("hex at i: " + string(hex[i]))
//		}
//	}
//
//	//currIndex=endIndex+1
//	//endIndex+=20
//	//make the eighth slice -- code line data
//	//str = hex[currIndex:endIndex]
//	//fmt.Println("Hex8: ", str)
//	//convert from hex to binary
//	codeLine := str
//	//set the value and print it
//	l4p.codeLineData = codeLine
//	fmt.Println("codeLine: ", codeLine)
//
//
//	//make the last slice -- CRC 16 check. Just need to store string version of the hex
//	currIndex=currIndex+1
//	str=""
//
//	for ; len(str) < 4; {
//		currChar = string(hex[currIndex])
//
//		if currChar == " "{
//			currIndex++
//		}else {
//			str += string(hex[currIndex])
//			currIndex++
//		}
//	}
//	//fmt.Println("Hex9: ", str)
//	//set and print the value
//	l4p.crc = str
//	fmt.Println("crc: ", str)
//
//	return l4p
//}

/* Utilities */

// Converts a 2 digit hex to a string of the bits
func HexToBinary(s string) string{
	if len(s) != 2 {
		return ""
	}
	//	Convert from hex to int
	dec, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	bitArray := fmt.Sprintf("%b", dec[0])
	//fmt.Println("bit array: ", bitArray)
	//fmt.Println("bit array: ", bits)
	if len(bitArray) < 8 {
		//fmt.Println("Length: ",len(bitArray))
		x:=8-len(bitArray)
		for i := 0; i < x; i++ {
			bitArray = "0" + bitArray
		}

	}
	//fmt.Println("bit array after: ", bitArray)
	//fmt.Println()
	return bitArray
}

/*
Input: Only pass in strings of size 2 ( 1 Hex at a time!)
 */
func HexToDec(s string) int {
	if len(s) == 1 {
		s = "0" + s
	}
	//	Convert from hex to int
	dec, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}

	return int(dec[0])

}
