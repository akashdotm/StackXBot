package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"net/http"
	"bytes"
	"encoding/json"
	"github.com/nlopes/slack"
)

//Stack Exchange Data.
type ownerDat struct {
	Reputation int `json:"reputation"`
	UserId int   `json:"user_id"`
	UserType string `json:"user_type"`
	ProfileImage string `json:"profile_image"`
	DisplayName string `json:"display_name"`
	Link string `json:"link"`
}


type stackXData struct {
	Tags []string `json:"tags"`
	Owner ownerDat `json:"owner"`
	IsAns	bool `json:"is_answered"`
	ViewCount int `json:"view_count"`
	AnswerCount int `json:"answer_count"`
	Score int `json:"score"`
	LastActivityDate int `json:"last_activity_date"`
	CreationDate int `json:"creation_date"`
	QuestionId int `json:"question_id"`
	Link string `json:"link"`
	Title string `json:"title"`
}

type stackXMain struct {
	Items []stackXData `json:"items"`
	HasMore bool `json:"has_more"`
	QuotaMax int `json:"quota_max"`
	QuotaRemain int `json:"quota_remaining"`
}

type userChannelData struct {
	CurrentPointInFlow string
	ProcessedText string
	MarkCurrentIterationComplete bool
	TitleContent string
	TagContent string
	UserDet string
	SiteName string
	IncomingMsg *slack.Msg 
}

type siteData struct {
	ApiSiteParameter string `json:"api_site_parameter"`
}

type stackSiteMain struct {
	Items []siteData `json:"items"`
	HasMore bool `json:"has_more"`
	QuotaMax int `json:"quota_max"`
	QuotaRemain int `json:"quota_remaining"`
}

var siteNames []siteData
var incomingReqUserChannelData []userChannelData
var welcomeMsg = []string{"hi","hello","whats up","hey","heyo","heya","yo"}
var	stackXKey = "KZWTIpTyVR*ptaBoN62v0A(("
var api = slack.New("xoxb-211211713988-OI4MrOVzkcwc3xshKhZFL6os")
func main() {
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(true)
	go fetchSiteNames(false) //fetch site names without meta sites [metaflag = false]
	rtm := api.NewRTM()
	go rtm.ManageConnection()
	var userAlreadyTalking bool
	for msg := range rtm.IncomingEvents {
		userAlreadyTalking = false
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace #general with your Channel ID
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "#general"))

		case *slack.MessageEvent:
		    
			fmt.Printf("Message: %v\n", ev)
			var currentMessageIndex int
			//var incomingMsg *slack.Msg
			//incomingMsg = &(ev.Msg)
			incomingReqUserChannel := userChannelData{IncomingMsg:&(ev.Msg),}
			for i,_ := range incomingReqUserChannelData {
				if incomingReqUserChannel.IncomingMsg.Channel == incomingReqUserChannelData[i].IncomingMsg.Channel {
					if incomingReqUserChannel.IncomingMsg.User == incomingReqUserChannelData[i].IncomingMsg.User{
					    fmt.Println("user details from the queue")
						fmt.Println(incomingReqUserChannelData[i])
						userAlreadyTalking = true
						incomingReqUserChannelData[i].IncomingMsg = &(ev.Msg)
						incomingReqUserChannel = incomingReqUserChannelData[i]
						currentMessageIndex = i
						incomingReqUserChannelData[i].MarkCurrentIterationComplete = false
					}
				}
			}
			if !userAlreadyTalking {
				incomingReqUserChannel.CurrentPointInFlow = "WELCOME"
				incomingReqUserChannel.MarkCurrentIterationComplete = false
				incomingReqUserChannelData = append(incomingReqUserChannelData,incomingReqUserChannel)
				currentMessageIndex = len(incomingReqUserChannelData)-1
				}
			fmt.Printf("Apna Message: %v\n", incomingReqUserChannel.IncomingMsg)
			go spawnCommunication(currentMessageIndex)
			
		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:

			// Ignore other events..
			// fmt.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}

func processText(mTitle string, mText string, siteName string,key string) string{
	searchTerms := strings.Split(mText," ")
	//var searchTerms []string
	//var maxTermSize int
	atleastOneResult := false
	//searchTerms[0] = strings.Join(setKeyword,"")
	//searchTerms[1] = strings.Join(setKeyword,"-")
	//n:=len(setKeyword)
	//fullWordRef := strings.Join(setKeyword[0:],"")
	//searchTerms = append(searchTerms,strings.Join(setKeyword[0:],""))
	//fullWordHyphenRef := strings.Join(setKeyword[0:],"-")
	maxTermSize := 40
	//searchTerms = append(searchTerms,strings.Join(setKeyword[0:],"-"))
	//for i:=1; i<n ; i++{
	//	searchTerms = append(searchTerms,strings.Join(setKeyword[:n-i-1],""))
	//	searchTerms = append(searchTerms,strings.Join(setKeyword[:n-i-1],"-"))	
	//	searchTerms = append(searchTerms,strings.Join(setKeyword[i:],""))
	//	searchTerms = append(searchTerms,strings.Join(setKeyword[i:],"-"))	
	//}
	fmt.Println(searchTerms)
	mTitle = strings.Replace(mTitle, " ", "%20", -1)
	var bufferForTags bytes.Buffer
	var stackResp []string
	var quesIdList []int
	var responseTitleTag []stackXData
	var responseTitle []stackXData
	siteApiParameter := predictSiteName(siteName)
	responseHeader := "Below are your responses: "
	bufferForTags.WriteString("&title=")
	bufferForTags.WriteString(mTitle)
	bufferForTags.WriteString("&site=")
	bufferForTags.WriteString(siteApiParameter)
	bufferForTags.WriteString("&key=")
	bufferForTags.WriteString(key)
	
	
	responseTitle = callStackApi(bufferForTags.String())				//TODO Change gotBack as array of StackXData, and materialize for loop through the items and adding to stackResp.

	if len(searchTerms[0])<maxTermSize {
		fmt.Println("printing tag")
		fmt.Println(searchTerms[0])
		
		bufferForTags.WriteString("&tagged=")
		bufferForTags.WriteString(searchTerms[0])
		responseTitleTag = callStackApi(bufferForTags.String())			//TODO Change gotBack as array of StackXData, and materialize for loop through the items and adding to stackResp.    
		}
    if len(responseTitleTag) > 0 {
		responseTitle = analyseResponse(responseTitleTag,responseTitle)
		}
	for _,eachReturnedResp := range responseTitle{
			if eachReturnedResp.Link !="NO_TOPICS" && !isItThere(quesIdList,eachReturnedResp.QuestionId) {
				if eachReturnedResp.Link != "DONT_INCLUDE" {
					if len(stackResp) == 0{
						stackResp = append(stackResp,responseHeader)
					}
					quesIdList = append(quesIdList,eachReturnedResp.QuestionId)
					//stackResp = append(stackResp,eachReturnedResp.Title)
					//stackResp = append(stackResp,": ")
					stackResp = append(stackResp,eachReturnedResp.Link)
					atleastOneResult = true
					} 
			}
		}
	bufferForTags.Reset()
	
	
	if atleastOneResult == false && len(stackResp)==0 {
		fmt.Println("inside atleastOneResult check")
		stackResp = append(stackResp,"Sorry! No discussions found for any keywords entered.")
	} else {
		//stackResp = append(responseHeader,stackResp)
	}
	
	responseSend := strings.Join(stackResp,"\n")
	return responseSend
	
}

func fetchSiteNames(metaFlag bool) {
	siteNames = nil     //reset the site name array
	url := "https://api.stackexchange.com/2.2/sites?pagesize=100"
	res, err := http.Get(url)
	if err != nil {
        	panic(err.Error())
    	}
	siteDec := json.NewDecoder(res.Body)
	var stackSite  = new(stackSiteMain) 
	err = siteDec.Decode(&stackSite)
	for _,item := range stackSite.Items {
		if strings.Contains(item.ApiSiteParameter,"meta") {
			if metaFlag {
				siteNames = append(siteNames,item)
			} else {
			}
		} else {
			siteNames = append(siteNames,item)
		}
	}
}

func predictSiteName(siteData string) string{
	var maxWeight int
	var currentWeight int
	var maxWtIndex int
	for i,item := range siteNames {
		fmt.Println(item.ApiSiteParameter)
		currentWeight = LCS(item.ApiSiteParameter,siteData)
		if(currentWeight > maxWeight){
			maxWeight = currentWeight
			maxWtIndex = i
			}
	}
	return siteNames[maxWtIndex].ApiSiteParameter
}

func callStackApi(tag string) []stackXData{
	
	//TODO change return type to []stackXData
	url := "https://api.stackexchange.com/2.2/similar?page=1&pagesize=5&order=desc&sort=relevance"
    var bufferParams bytes.Buffer
	bufferParams.WriteString(url)
	bufferParams.WriteString(tag)
	requestURL := bufferParams.String()
	fmt.Println("Printing entire URL, below")
	fmt.Println(requestURL)
	res, err := http.Get(requestURL)
	if err != nil {
        	panic(err.Error())
    	}
	//defer res.Body.Close()
	//body, err := ioutil.ReadAll(res.Body)
	dec := json.NewDecoder(res.Body)
	var stack1  = new(stackXMain) 
	//var stack1 stackXMain
	err = dec.Decode(&stack1)
	//err1 := json.Unmarshal(body, &stack1 )
	if err!=nil{
		panic(err)
	}
	for _,item := range stack1.Items {
		fmt.Println(item.Title)
	}
	if len(stack1.Items) == 0 {
			var dummyStack []stackXData
			stack1 := stackXData{QuestionId:0,Link:"NO_TOPICS",Title:"Oops! No topics found for the above tags."}
			dummyStack = append(dummyStack,stack1)
			fmt.Println("No topics found for %s",tag)
			bufferParams.Reset()
			return dummyStack
		} else {
			fmt.Println(stack1.Items[0].Title)
			//var returnItems bytes.Buffer
			//returnItems.WriteString(stack1.Items[0].Title)
			//returnItems.WriteString(": ")
			//returnItems.WriteString(stack1.Items[0].Link)
			//bufferParams.Reset()
			return stack1.Items    //TODO Change to stack1.Items
			}
} 

func isItThere(arr []int, num int) bool {
   for _, a := range arr {
      if a == num {
         return true
      }
   }
   return false
}

func analyseResponse(xone []stackXData, xtwo []stackXData) []stackXData{
	// xone,xtwo length[wt: 2]; viewcount for both [wt: 2]; answercount | score for each [wt: 1-1 each] 
	var wt_one int
	var wt_two int
	var score_one int
	var score_two int
	var answercount_one int
	var answercount_two int
	sizeone := len(xone)
	sizetwo := len(xtwo)
	var viewcntone int
	var viewcnttwo int
	
	if sizeone == sizetwo {
	} 
	if sizeone > sizetwo {
		wt_one += 2
	}
	if sizetwo > sizeone {
		wt_two += 2
	}
	
	for _,each_one := range xone {
		viewcntone += each_one.ViewCount
		answercount_one += each_one.AnswerCount
		score_one += each_one.Score
	}
	for _,each_two := range xtwo {
		viewcnttwo += each_two.ViewCount
		answercount_two += each_two.AnswerCount
		score_two += each_two.Score
	}
	
	if viewcntone == viewcnttwo {
	} 
	if viewcntone > viewcnttwo {
		wt_one += 2
	}
	if viewcnttwo > viewcntone {
		wt_two += 2
	}
	
	if answercount_one == answercount_two {
	}
	if answercount_one > answercount_two {
		wt_one +=1
	}
	if answercount_one < answercount_two {
		wt_two += 1
	}
	
	if score_one == score_two {
	}
	if score_one > score_two {
		wt_one += 1
	}
	if score_one<score_two {
		wt_two += 1
		}
	//TODO answercount and score
	if wt_one >= wt_two {
		return xone
		} else {
		return xtwo
		}
	
}

func Max(more ...int) int {
	max_num := more[0]
	for _, elem := range more {
		if max_num < elem {
			max_num = elem
		}
	}
	return max_num
}

func LCS(str1, str2 string) int {
	len1 := len(str1)
	len2 := len(str2)

	table := make([][]int, len1+1)
	for i := range table {
		table[i] = make([]int, len2+1)
	}

	i, j := 0, 0
	for i = 0; i <= len1; i++ {
		for j = 0; j <= len2; j++ {
			if i == 0 || j == 0 {
				table[i][j] = 0
			} else if str1[i-1] == str2[j-1] {
				table[i][j] = table[i-1][j-1] + 1
			} else {
				table[i][j] = Max(table[i-1][j], table[i][j-1])
			}
		}
	}
	return table[len1][len2]
}

func spawnCommunication(currentMessageIndex int) {
	
	
	var incomingMsg *slack.Msg
	incomingMsg = incomingReqUserChannelData[currentMessageIndex].IncomingMsg
	fmt.Println("Entering conversation")
	fmt.Println("Message as follows")
	fmt.Println(incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow)
	fmt.Println("-----------------------------------------")
	fmt.Println(incomingMsg.User)
	fmt.Println("-----------------------------------------")
	fmt.Println(incomingMsg.Text)
	//Welcome message 
	if incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow == "WELCOME" && !incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete {
			fmt.Println("--------------------inside welcome ----------------------------------")
			//if errUser != nil {
			//	fmt.Printf("%s\n", errUser)
			//	return
			//}
			for wc := range welcomeMsg{
				if strings.HasPrefix(strings.ToLower(incomingMsg.Text),welcomeMsg[wc]) {
					welcomeText := "Howdy! What are the queries that I can help you with today?"
					params := slack.PostMessageParameters{}
					_, _, err := api.PostMessage(incomingMsg.Channel, welcomeText, params)
					if err != nil {
						fmt.Printf("%s\n", err)
						return
					}
					incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow = "GET_TITLE"
					incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete = true
					break
				}
			}	
		}
			
	//Fetch Title
	if incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow == "GET_TITLE" && !incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete {
		fmt.Println("--------------------inside get title ----------------------------------")
		attachInfo := slack.Attachment{
					Pretext: "Please choose the appropriate sort option by clicking on the desirable button",
					Text: "Just click on any of the buttons, to choose a sort method for your responses",
					CallbackID: "SortTypeButtons",
					Fallback:"Unable to choose a method",
					Color:"#3AA3E3",
					Title:"Choose any sort option",
					}
		incomingReqUserChannelData[currentMessageIndex].TitleContent = incomingMsg.Text 
		afterTitleText := "If you want any relevant tags for the query, enter here [each tag separated by space or comma]. If not, enter nothing."
		params := slack.PostMessageParameters{}
		params.Attachments = []slack.Attachment{attachInfo}
		_, _, err := api.PostMessage(incomingMsg.Channel, afterTitleText, params)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
		fmt.Println(afterTitleText)
		incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow = "GET_TAGS"
		incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete = true
	}
		
	//Fetch Tags
	if incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow == "GET_TAGS" && !incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete {
			fmt.Println("--------------------inside get tags ----------------------------------")
			if len(incomingMsg.Text)<70 && incomingMsg.Text!="" {
				
				incomingReqUserChannelData[currentMessageIndex].TagContent = incomingMsg.Text
				incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete = true
				incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow = "GET_SITE"
			}
		afterSiteText := "Which Stack exchange you want to interact with?"
		params := slack.PostMessageParameters{}
		_, _, err := api.PostMessage(incomingMsg.Channel, afterSiteText, params)
		if err != nil {
			fmt.Printf("%s\n", err)
			return
		}
	}
	
	//Fetch site name
	if incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow == "GET_SITE" && !incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete {
		fmt.Println("--------------------inside get site ----------------------------------")
		incomingReqUserChannelData[currentMessageIndex].SiteName = incomingMsg.Text
		incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow = "PROCESSING"
		incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete = true
	}
			
	//Put up text for processing.
	if incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow == "PROCESSING" {
			fmt.Println("--------------------inside PROCESSING----------------------------------")
			if strings.Contains(incomingMsg.Text,"meta") {
				go fetchSiteNames(true)
			}
			incomingReqUserChannelData[currentMessageIndex].ProcessedText = processText(incomingReqUserChannelData[currentMessageIndex].TitleContent,incomingReqUserChannelData[currentMessageIndex].TagContent,incomingReqUserChannelData[currentMessageIndex].SiteName,stackXKey)
			params := slack.PostMessageParameters{}
			_, _, err := api.PostMessage(incomingMsg.Channel, incomingReqUserChannelData[currentMessageIndex].ProcessedText, params)
			if err != nil {
				fmt.Printf("%s\n", err)
				return
			}
			
			incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow = "RESET"
		}
			
	//Reset the iteration.
	if incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow == "RESET" && !incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete  {
			fmt.Println("--------------------inside reset ----------------------------------")
			if incomingMsg.Text == "yes" || incomingMsg.Text == "Yes"{
				incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow = "WELCOME"
			} else {
				incomingReqUserChannelData[currentMessageIndex].CurrentPointInFlow = "END"
				// remove user channel data from incomingReqUserChannelData
				//for removeIdx := currentMessageIndex + 1,removeIdx<len(incomingReqUserChannelData),removeIdx++ {
				//	incomingReqUserChannelData[removeIdx-1] = incomingReqUserChannelData[removeIdx]
				//}
			}
			incomingReqUserChannelData[currentMessageIndex].MarkCurrentIterationComplete = true
		}
}