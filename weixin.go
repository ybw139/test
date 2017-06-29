package controllers
//test1
//test2
import (
	"fmt"
	"strconv"

	"github.com/chanxuehong/wechat.v2/mp/core"
	"github.com/chanxuehong/wechat.v2/mp/material"
	"github.com/chanxuehong/wechat.v2/mp/menu"
	"github.com/chanxuehong/wechat.v2/mp/message/mass/mass2group"
	"github.com/chanxuehong/wechat.v2/mp/user/group"
	"nxstory.com/models"
	"runtime"
)

type WeixinController struct {
	CommonController
}

func GroupDown(AppId, AppSecret string) {

	wechatClient := core.NewClient(core.NewDefaultAccessTokenServer(AppId, AppSecret, nil), nil)
	male_id, female_id, unknown_id := checkgroup(AppId, wechatClient)
	_, err := models.GetCsGroupByAppGroupId(AppId, int(male_id))
	if err != nil && male_id > 0 {
		var _group models.CsGroup
		_group.WxAppId = AppId
		_group.Name = "男性用户"
		_group.Group = int(male_id)
		models.AddCsGroup(&_group)
	}
	_, err = models.GetCsGroupByAppGroupId(AppId, int(female_id))
	if err != nil && female_id > 0 {
		var _group models.CsGroup
		_group.WxAppId = AppId
		_group.Name = "女性用户"
		_group.Group = int(female_id)
		models.AddCsGroup(&_group)
	}
	_, err = models.GetCsGroupByAppGroupId(AppId, int(unknown_id))
	if err != nil && unknown_id > 0 {
		var _group models.CsGroup
		_group.WxAppId = AppId
		_group.Name = "未知用户"
		_group.Group = int(unknown_id)
		models.AddCsGroup(&_group)
	}
	fmt.Println("maleid:", male_id, "femaleid:", female_id, "unknownid:", unknown_id)

}
func checkgroup(AppId string, wechatClient *core.Client) (int64, int64, int64) {
	female_id := int64(0)  //女性用户分类
	male_id := int64(0)    //男性用户分类
	unknown_id := int64(0) //未知用户分类
	wxgroup, _ := group.List(wechatClient)
	for _, g := range wxgroup {
		if g.Name == "男性用户" {
			male_id = g.Id
		}
		if g.Name == "女性用户" {
			female_id = g.Id
		}
		if g.Name == "未知用户" {
			unknown_id = g.Id
		}
		_, err := models.GetCsGroupByAppGroupId(AppId, int(g.Id))
		if err != nil && unknown_id > 0 {
			var _group models.CsGroup
			_group.WxAppId = AppId
			_group.Name = g.Name
			_group.Group = int(g.Id)
			models.AddCsGroup(&_group)
		}
	}
	if male_id == 0 {
		g, _ := group.Create(wechatClient, "男性用户")
		male_id = g.Id
	}
	if female_id == 0 {
		g, _ := group.Create(wechatClient, "女性用户")
		female_id = g.Id
	}
	if unknown_id == 0 {
		g, _ := group.Create(wechatClient, "未知用户")
		unknown_id = g.Id
	}
	return male_id, female_id, unknown_id
}
func MaterialSend(AppId, AppSecret, MediaId, GroupId string, ToAll bool) {

	msg := make(map[string]interface{})
	msg_filter := make(map[string]interface{})
	msg_news := make(map[string]string)
	msg_news["media_id"] = MediaId
	msg_filter["is_to_all"] = ToAll
	if !ToAll {
		msg_filter["group_id"] = GroupId
	}

	msg["filter"] = msg_filter
	msg["mpnews"] = msg_news
	msg["msgtype"] = "mpnews"
	fmt.Println(msg)

	wechatClient := core.NewClient(core.NewDefaultAccessTokenServer(AppId, AppSecret, nil), nil)

	_, err := mass2group.Send(wechatClient, msg)
	if err != nil {
		fmt.Println(err.Error())
	}

}
func checkMaterial(wxclient *core.Client, AppId string) {
	limit := 20
	from := 0
PROCESS:
	news, err := material.BatchGetNews(wxclient, from, limit)
	if err == nil {
		fmt.Println(news.ItemCount, news.TotalCount)
		for k, v := range news.Items {
			fmt.Println(k, v.MediaId, v.UpdateTime)
			m, err1 := models.GetCsMaterialByAppMId(AppId, v.MediaId)
			fmt.Println(err1)
			if err1 == nil {
				m.Content = v.Content.Articles[0].Title
				models.UpdateCsMaterialById(m)
			} else {
				var _m models.CsMaterial
				_m.WxAppId = AppId
				_m.Mid = v.MediaId
				_m.Mtype = "mpnews"
				_m.Content = v.Content.Articles[0].Title
				models.AddCsMaterial(&_m)
			}
		}
		if news.ItemCount == limit {
			from = from + limit
			fmt.Println("from:", from)
			goto PROCESS
		}
	}
}
func (c *WeixinController) Get() {
	c.CheckLogin()
	query := map[string]string{}

	fields := []string{"Id", "WxAppId", "Name", "Mail"}
	sortby := []string{"Id"}
	order := []string{"desc"}

	items, _ := models.GetAllCsWxAccount(query, fields, sortby, order, 0, 200)

	c.Data["items"] = items
	c.Data["Title"] = "帐号管理"
	c.Data["TitleSmall"] = "列表"
	c.TplName = "weixin/index.html"
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["HeaderExt"] = "headerext_list.html"
	c.LayoutSections["FooterExt"] = "footerext_list.html"
}

func (c *WeixinController) MenuEdit() {
	c.CheckLogin()
	appid := c.Ctx.Input.Param(":appid")
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	typeStr := c.Ctx.Input.Param(":type")
	typenum, _ := strconv.Atoi(typeStr)
	if id != 0 && typenum != 2 {
		item, _ := models.GetCsWxMenuById(id)
		if item != nil {
			c.Data["item"] = item
		}else{
			c.Data["item"] = models.CsWxMenu{}
		}
		c.TplName = "weixin/edit.html"
		c.Data["Title"] = "编辑菜单"
		c.Data["TitleSmall"] = "编辑"
	} else {
		query := map[string]string{}
		if idStr != "" {
			query["pid"] = idStr
		}
		fields := []string{"Id", "WxAppId", "OrderId","Name", "Type", "Pid", "Eventtype", "Tourl"}
		sortby := []string{"OrderId"}
		order := []string{"asc"}

		items, _ := models.GetAllCsWxMenu(query, fields, sortby, order, 0, 200)
		c.Data["pid"] = id
		c.Data["items"] = items
		c.Data["Title"] = "管理二级菜单"
		c.Data["TitleSmall"] = "列表"
		c.TplName = "weixin/menu.html"
	}
	c.Data["wxappid"] = appid
	c.Data["flag"] = typenum
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["HeaderExt"] = "headerext_list.html"
	c.LayoutSections["FooterExt"] = "footerext_list.html"
}

func (c *WeixinController) MenuAdd() {
	c.CheckLogin()
	appid := c.Ctx.Input.Param(":appid")
	item := models.CsWxMenu{}
	c.Data["wxappid"] = appid
	c.Data["flag"] = 0
	c.Data["item"] = item
	c.Data["Title"] = "添加二级菜单"
	c.Data["TitleSmall"] = "添加"
	c.TplName = "weixin/edit.html"
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["HeaderExt"] = "headerext_edit.html"
	c.LayoutSections["FooterExt"] = "footerext_edit.html"
}

func (c *WeixinController) MenuAddDo() {
	c.CheckLogin()
	appid := c.Ctx.Input.Param(":appid")
	idStr := c.Ctx.Input.Param(":pid")
	pid, _ := strconv.Atoi(idStr)

	item := models.CsWxMenu{}
	item.Pid = pid
	if c.GetString("Name") != "" {
		item.Name = c.GetString("Name")
	}
	if val, _ := c.GetInt("Type"); val != 0 {
		item.Type, _ = c.GetInt("Type")
	}
	if val, _ := c.GetInt("OrderId"); val != 0 {
		item.OrderId, _ = c.GetInt("OrderId")
	}
	if c.GetString("Tourl") != "" {
		item.Tourl = c.GetString("Tourl")
	}
	if c.GetString("Eventtype") != "" {
		item.Eventtype = c.GetString("Eventtype")
	}
	models.AddCsWxMenu(&item)
	c.Redirect("/weixin/menu/"+appid, 302)
}

func (c *WeixinController) MenuEditDo() {
	c.CheckLogin()
	appid := c.Ctx.Input.Param(":appid")
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	//取得书信息
	item, _ := models.GetCsWxMenuById(id)
	if item != nil {
		if c.GetString("Name") != "" {
			item.Name = c.GetString("Name")
		}
		if val, _ := c.GetInt("Type"); val != 0 {
			item.Type, _ = c.GetInt("Type")
		}
		if val, _ := c.GetInt("OrderId"); val != 0 {
			item.OrderId, _ = c.GetInt("OrderId")
		}
		if c.GetString("Tourl") != "" {
			item.Tourl = c.GetString("Tourl")
		}
		if c.GetString("Eventtype") != "" {
			item.Eventtype = c.GetString("Eventtype")
		}

		models.UpdateCsWxMenuById(item)
	}
	c.Redirect("/weixin/menu/"+appid, 302)
}

func (c *WeixinController) MenuDel()  {
	c.CheckLogin()
	appid := c.Ctx.Input.Param(":appid")
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	models.DeleteCsWxMenu(id)
	c.Redirect("/weixin/menu/"+appid, 302)
}

func (c *WeixinController) Menu() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":appid")
	query := map[string]string{}
	if idStr != "" {
		query["wxappid"] = idStr
	}
	fields := []string{"Id", "WxAppId","OrderId", "Name", "Type", "Pid", "Eventtype", "Tourl"}
	sortby := []string{"OrderId"}
	order := []string{"asc"}

	items, _ := models.GetAllCsWxMenu(query, fields, sortby, order, 0, 200)
	if items == nil {
		item := models.CsWxMenu{}
		item.Pid = 0
		item.Type = 1
		item.OrderId =1
		item.WxAppId = idStr
		item.Name = "一级菜单名称"
		models.AddCsWxMenu(&item)
		item1 := models.CsWxMenu{}
		item1.Pid = 0
		item1.Type = 1
		item1.OrderId =2
		item1.WxAppId = idStr
		item1.Name = "一级菜单名称"
		models.AddCsWxMenu(&item1)
		item2 := models.CsWxMenu{}
		item2.Pid = 0
		item2.Type = 1
		item2.OrderId =3
		item2.WxAppId = idStr
		item2.Name = "一级菜单名称"
		models.AddCsWxMenu(&item2)
		fmt.Println(item)
		items, _ := models.GetAllCsWxMenu(query, fields, sortby, order, 0, 200)
		c.Data["items"] = items
	}else {
		c.Data["items"] = items
	}
	c.Data["wxappid"] = idStr
	c.Data["Title"] = "菜单管理"
	c.Data["TitleSmall"] = "列表"
	c.TplName = "weixin/menu.html"
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["HeaderExt"] = "headerext_list.html"
	c.LayoutSections["FooterExt"] = "footerext_list.html"
}
var appmap  map[string]*core.Client = make(map[string]*core.Client)

func (c *WeixinController) MenuSend() {
	c.CheckLogin()
	fmt.Println(runtime.NumGoroutine())
	idStr := c.Ctx.Input.Param(":appid")
	channel, err := models.GetCsWxAccountByAppId(idStr)
	wechatClient, ok := appmap[idStr]
	if !ok {
		wechatClient  = core.NewClient(core.NewDefaultAccessTokenServer(channel.WxAppId, channel.AppSecret, nil), nil)
		appmap[idStr] = wechatClient
	}
	fmt.Println(appmap)
	//info, bmenu, err := menu.GetMenuInfo(wechatClient)
	//if err != nil {
	//	fmt.Println("err:", err.Error())
	//} else {
	//	fmt.Println("bmenu:", bmenu)
	//	fmt.Println("button:", info.Buttons)
	//}

	query := map[string]string{}
	if idStr != "" {
		query["wxappid"] = idStr
	}
	fields := []string{"Id", "WxAppId", "Name", "OrderId","Type", "Pid", "Eventtype", "Tourl"}
	sortby := []string{"OrderId"}
	order := []string{"asc"}
	m := menu.Menu{}
	button := menu.Button{}
	items, _ := models.GetAllCsWxMenu(query, fields, sortby, order, 0, 200)
	if items != nil {
		for _, value := range items {
			item := value.(map[string]interface{})
			pid := item["Id"].(int)
			name := item["Name"].(string)
			url := item["Tourl"].(string)
			key := item["Eventtype"].(string)
			if item["Type"].(int) == 1 {
				//跳转
				button.SetAsViewButton(name, url)
				m.Buttons = append(m.Buttons, button)
			} else if item["Type"].(int) == 3 {
				//click事件
				button.SetAsClickButton(name, key)
				m.Buttons = append(m.Buttons, button)

			} else if item["Type"].(int) == 2 {
				//二级菜单
				subbutton := []menu.Button{}

				strconv.Itoa(pid)
				query := map[string]string{}
				query["pid"] = strconv.Itoa(pid)

				fields := []string{"Id", "WxAppId","OrderId", "Name", "Type", "Pid", "Eventtype", "Tourl"}
				sortby := []string{"OrderId"}
				order := []string{"asc"}

				itemsMore, _ := models.GetAllCsWxMenu(query, fields, sortby, order, 0, 200)

				if items != nil {
					for _, val := range itemsMore {
						item := val.(map[string]interface{})
						name := item["Name"].(string)
						url := item["Tourl"].(string)
						key := item["Eventtype"].(string)

						if item["Type"].(int) == 1 {
							//跳转
							button.SetAsViewButton(name, url)
							subbutton = append(subbutton, button)
						} else if item["Type"].(int) == 3 {
							//click事件
							button.SetAsClickButton(name, key)
							subbutton = append(subbutton, button)

						}

					}
				}

				button.SetAsSubMenuButton(name, subbutton)
				m.Buttons = append(m.Buttons, button)

			}

		}
	}
	err = menu.Create(wechatClient, &m)
	fmt.Println("createmenu:", err)
	c.Redirect("/weixin", 302)
}

func (c *WeixinController) Material() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":appid")

	query := map[string]string{}
	if idStr != "" {
		query["wxappid"] = idStr
	}
	fmt.Println("query:", query)
	fields := []string{"Id", "WxAppId", "Mid", "Mtype", "Content"}
	sortby := []string{"Id"}
	order := []string{"desc"}

	items, _ := models.GetAllCsMaterial(query, fields, sortby, order, 0, 200)

	c.Data["items"] = items

	c.Data["wxappid"] = idStr
	c.Data["Title"] = "素材管理"
	c.Data["TitleSmall"] = "素材下载"
	c.TplName = "weixin/material.html"
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["HeaderExt"] = "headerext_list.html"
	c.LayoutSections["FooterExt"] = "footerext_list.html"
}
func (c *WeixinController) MaterialDown() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":appid")
	channel, err := models.GetCsWxAccountByAppId(idStr)
	if err == nil {
		wechatClient := core.NewClient(core.NewDefaultAccessTokenServer(channel.WxAppId, channel.AppSecret, nil), nil)
		checkMaterial(wechatClient, channel.WxAppId)

	}
	c.Redirect("/weixin", 302)
}
func (c *WeixinController) GroupDown() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":appid")
	channel, err := models.GetCsWxAccountByAppId(idStr)
	if err == nil {
		GroupDown(channel.WxAppId, channel.AppSecret)
	}
	c.Redirect("/weixin", 302)
}
func (c *WeixinController) SendGroup() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	//取得书信息
	item, _ := models.GetCsMaterialById(id)
	if item != nil {
		c.Data["item"] = item
		query := map[string]string{}
		query["wxappid"] = item.WxAppId
		fields := []string{"Id", "WxAppId", "Name", "Group"}
		sortby := []string{"Id"}
		order := []string{"desc"}
		groups, _ := models.GetAllCsGroup(query, fields, sortby, order, 0, 200)
		c.Data["groups"] = groups
	}

	c.Data["Title"] = "消息发送"
	c.Data["TitleSmall"] = "消息发送"
	c.TplName = "weixin/send.html"
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["HeaderExt"] = "headerext_edit.html"
	c.LayoutSections["FooterExt"] = "footerext_edit.html"

}
func (c *WeixinController) SendGroupDo() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":id")
	groupid := c.GetString("groupid")
	id, _ := strconv.Atoi(idStr)
	//取得书信息
	item, err := models.GetCsMaterialById(id)
	if err == nil && groupid != "" {
		channel, _ := models.GetCsWxAccountByAppId(item.WxAppId)

		fmt.Println(channel.WxAppId, channel.AppSecret, item.Mid, groupid, false)
		MaterialSend(channel.WxAppId, channel.AppSecret, item.Mid, groupid, false)

	}
	c.Redirect("/weixin", 302)

}
func (c *WeixinController) SendAll() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	//取得书信息
	item, _ := models.GetCsMaterialById(id)
	if item != nil {
		c.Data["item"] = item

	}

	c.Data["Title"] = "消息发送"
	c.Data["TitleSmall"] = "消息发送"
	c.TplName = "weixin/send.html"
	c.Layout = "layout.html"
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["HeaderExt"] = "headerext_edit.html"
	c.LayoutSections["FooterExt"] = "footerext_edit.html"

}

func (c *WeixinController) SendAllDo() {
	c.CheckLogin()
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	//取得书信息
	item, err := models.GetCsMaterialById(id)
	if err == nil {
		channel, _ := models.GetCsWxAccountByAppId(item.WxAppId)

		fmt.Println(channel.WxAppId, channel.AppSecret, item.Mid, "", true)
		MaterialSend(channel.WxAppId, channel.AppSecret, item.Mid, "", true)

	}
	c.Redirect("/weixin", 302)

}
