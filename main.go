package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

func main(){
	// 기존 파일 제거
	os.Remove("lotto.csv")

	channel :=make(chan []string)
	results := [][]string{} // 크롤링한 각 회차의 로또 번호 배열로 저장

	numbers := getAllRoundNumber() // 가져올 수 있는 회차 목록

	// 각 회차의 번호 가져오는 goroutine 실행
	for i := 0; i < len(numbers); i++ {
		go getLottoNum(numbers[i], channel)
	}

	// 실행을 마친 goroutine 결과 값 저장
	for i := 0; i < len(numbers); i++ {
		result := <-channel
		results = append(results, result);
	}

	// results 회차 순서로 정렬
	sort.Slice(results, func(i, j int) bool {
		v1,_ := strconv.Atoi(results[i][0])
		v2,_ := strconv.Atoi(results[j][0])

		return v1 < v2 
})

	makeCSV(results)
}

// 크롤링할 수 있는 회차 목록을 반환
func getAllRoundNumber()[]string{
	res,err:=http.Get("https://search.daum.net/search?q=로또+당첨+번호")
	checkErr(err)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	
	searchCards := doc.Find("select.opt_select").Find("option")
	
	numbers := []string{}

	searchCards.Each(func(i int, card *goquery.Selection) {
		if i==0{
			return
		}

		sampleRegexp := regexp.MustCompile("(,)|((회차).+)")
		numbers = append(numbers, sampleRegexp.ReplaceAllString(card.Text(),""))
	})

	return numbers
}

func checkErr(err error){
	if err != nil{
		log.Fatalln(err)
	}
}


// 크롤링한 데이터를 csv파일로 저장
func makeCSV(results [][]string) {
	file, err := os.Create("lotto.csv")
	checkErr(err)
	defer file.Close() // 함수 종료 전 파일 닫기

	w := csv.NewWriter(file)
	defer w.Flush() // 버퍼에 담은 내용을 파일에 작성

	headers := []string{"round","num1","num2","num3","num4","num5","num6", "bonus"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, result  := range results{
		jobSlice := result
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}

}

// 특정 회차의 번호를 크롤링
func getLottoNum(num string, channel chan<- []string){
	res, err:= http.Get("https://search.daum.net/search?&q="+num+"회차%20로또")
	checkErr(err)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find("div.lottonum").Find("span.ball")

	ns := []string{num}

	searchCards.Each(func(i int, card *goquery.Selection) {
		numTxt := card.Text()

		// span 태그 중 + 기호를 의미하는 span 제거
		if numTxt != "보너스"{
			ns = append(ns, numTxt)
		}
	})

	channel<- ns
}