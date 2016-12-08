package module

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/zieckey/simgo"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var (
	ErrConvert = errors.New("convert error")
	Sqp        *SearchRequestParams
	mu         sync.Mutex
)

//解析查询参数
type SearchRequestParams struct {
	Days        []interface{}          `json:"days"`
	Page_size   int32                  `json:"page_size"`
	Page_number int32                  `json:"page_number"`
	Day         string                 `json:"day"`
	Business    string                 `json:"business"`
	Keywords    map[string]interface{} `json:"keywords"`
	Options     map[string]interface{} `json:"options"`
	//	Filters     map[string]interface{} `json:"filters"`
}

func NewSearchRequestParams() *SearchRequestParams {
	return &SearchRequestParams{
		Days:     make([]interface{}, 0),
		Keywords: make(map[string]interface{}),
		Options:  make(map[string]interface{}),
		//		Filters:  make(map[string]interface{}),
	}
}

type QueryBody struct {
	Query *SearchRequestParams `json:"query"`
}

func NewQuery() *QueryBody {
	return &QueryBody{
		Query: NewSearchRequestParams(),
	}
}

type Proxy struct {
	poseidon_search_url string
}

func New() *Proxy {
	return &Proxy{}
}

func (p *Proxy) Initialize() error {
	fw := simgo.DefaultFramework
	p.poseidon_search_url, _ = fw.Conf.SectionGet("proxy", "poseidon_search_url")

	simgo.HandleFunc("/service/proxy/mdsearch", p.MdsearchAction, p).Methods("POST")
	Sqp = NewSearchRequestParams()
	return nil
}

func (p *Proxy) Uninitialize() error {
	return nil
}

/**
 * multi day search
 * @param  {[type]} this *SearchController) MdsearchAction( [description]
 * @return {[type]}      [description]
 */
func (this *Proxy) MdsearchAction(w http.ResponseWriter, r *http.Request) {
	days, err := this.GetDays(r)
	if err != nil {
		panic(err)
	}
	tasknum := len(days)
	log.Println("tasknum:", tasknum, days)
	log.Println(days)
	//init result channel container
	c := make(chan string, tasknum)

	for _, day := range days {
		if day == "" {
			continue
		}
		go this.send(day, c)
	}

	//recieve result
	//var response_num string
	buf := bytes.NewBuffer([]byte("["))
	for i := 0; i < tasknum; i++ {
		chanr := <-c
		buf.WriteString(chanr)
		if i != (tasknum - 1) {
			buf.WriteString(",")
		}
	}
	buf.WriteString("]")
	w.Write(buf.Bytes())
}

/**
 * send request put data into channel
 * @param  {[type]} this *SearchController) send(day string, c chan string [description]
 * @return {[type]}      [description]
 */
func (this *Proxy) send(day string, c chan string) {
	defer func() {
		if err := recover(); err != nil {
			c <- "request timeout"
		}
	}()
	b, _ := this.GetPostBody(day)
	body := bytes.NewBuffer(b)
	req, err := http.NewRequest("POST", this.poseidon_search_url, body)
	log.Println("send url ", this.poseidon_search_url)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	re, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	c <- string(re)
}

func (this *Proxy) getparams(r *http.Request) (*SearchRequestParams, error) {
	jsonenc := json.NewDecoder(r.Body)
	searchParams := make(map[string]interface{}, 1000)
	err := jsonenc.Decode(&searchParams)
	if err != nil {
		return nil, err
	}

	query, ok := searchParams["query"].(map[string]interface{})
	if !ok {
		return nil, ErrConvert
	}

	Sqp.Page_size = int32(query["page_size"].(float64))
	Sqp.Page_number = int32(query["page_number"].(float64))
	Sqp.Business = query["business"].(string)
	Sqp.Keywords = query["keywords"].(map[string]interface{})

	Sqp.Options = query["options"].(map[string]interface{})
	//Sqp.Filters = query["filters"].(map[string]interface{})

	if query["day"] != nil {
		Sqp.Day = query["day"].(string)
	}
	if query["days"] != nil {
		Sqp.Days = query["days"].([]interface{})
	}

	return Sqp, nil
}

func (this *Proxy) GetDays(r *http.Request) ([]string, error) {
	params, err := this.getparams(r)
	if err != nil {
		return nil, err
	}

	//初始化新容器，断离params大对象
	days := make([]string, len(params.Days))
	for i, day := range params.Days {
		if newday, ok := day.(string); ok {
			days[i] = newday
		}
	}

	return days, nil
}

func (this *Proxy) GetPostBody(day string) ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	Sqp.Day = day
	query := NewQuery()
	query.Query = Sqp
	body, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	return body, nil
}
