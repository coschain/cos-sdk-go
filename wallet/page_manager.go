package wallet

type Page struct {
	Start interface{}
	End interface{}
	Limit uint32
	LastOrder interface{}
}

type PageManager struct {
	pageList []*Page
	callFunc CallRpc
	pageIndex uint32
}
type CallRpc func(p *PageManager) interface{}

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
	pm.pageIndex = 0
	return pm
}

func (pm *PageManager) Next() interface{} {
	res := pm.callFunc(pm)
	return res
}

func (pm *PageManager) CurrentPage() uint32 {
	return pm.pageIndex
}

func (pm *PageManager) PageCount() uint32 {
	return uint32(len(pm.pageList))
}