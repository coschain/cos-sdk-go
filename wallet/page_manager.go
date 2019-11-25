package wallet

import "github.com/kataras/go-errors"

type Page struct {
	Start interface{}
	End interface{}
	Limit uint32
	LastOrder interface{}
}

type PageManager struct {
	pageList    []*Page
	callFunc    CallRpc
	currentPage int
	end         interface{}
}

// callback function for execute rpc call
// return: rpc result
// return: lastOrder of result
// return: error msg
type CallRpc func(page *Page) (interface{},interface{},error)

func NewPageManager(start interface{}, end interface{}, limit uint32, lastOrder interface{}, rpcCallBack CallRpc) *PageManager {
	pm := &PageManager{}
	p := &Page{
		Start:start,
		End:end,
		Limit:limit,
		LastOrder:lastOrder,
	}
	pm.pageList = append(pm.pageList,p)
	pm.callFunc = rpcCallBack
	pm.currentPage = -1
	pm.end = end
	return pm
}

func (pm *PageManager) Next() (interface{},error) {
	// find page
	requestPage := pm.CurrentPage()+1
	page := pm.GetPage(requestPage)
	if page == nil {
		return nil,errors.New("page out of range")
	}

	// call rpc
	res,lastOrder,err := pm.callFunc(page)
	if err != nil {
		return nil,err
	}

	// add new page for next query
	if requestPage + 1 == pm.PageCount() {
		pm.AddPage(&Page{Start: lastOrder, End: pm.end, Limit: page.Limit, LastOrder: lastOrder})
	}
	pm.SetCurrentPage(requestPage)

	return res,nil
}

func (pm *PageManager) Pre() (interface{},error) {
	requestPage := pm.CurrentPage()-1
	page := pm.GetPage(requestPage)
	if page == nil {
		return nil,errors.New("page out of range")
	}

	res, _, err := pm.callFunc(page)
	if err != nil {
		return nil,err
	}

	// todo trim last page

	pm.SetCurrentPage(requestPage)

	return res, nil
}

func (pm *PageManager) CurrentPage() int {
	return pm.currentPage
}

func (pm *PageManager) PageCount() int {
	return len(pm.pageList)
}

func (pm *PageManager) SetCurrentPage(page int) {
	pm.currentPage = page
}

func (pm *PageManager) AddPage(p *Page) {
	pm.pageList = append(pm.pageList, p)
}

func (pm *PageManager) GetPage(index int) *Page {
	if index < 0 || index >= pm.PageCount() {
		return nil
	}
	return pm.pageList[index]
}