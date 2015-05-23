//import "/kb/KB.js"
//import "/kb/Lineup.View.js"

KB.Site = (function(){
	"use strict"

	var HeaderMenu = React.createClass({
		displayName: "HeaderMenu",
		render: function(){
			return React.DOM.div({
				className:"header-menu"
			},
				this.props.items.map(function(item, i){
					return React.DOM.a(item, item.caption);
				})
			);
		}
	})

	var Header = React.createClass({
		displayName: "Header",
		render: function(){
			var a = React.DOM.a;
			return React.DOM.div({
				id:"header"
			},
				a({className:"button logo", href:"/", title:"Home"}),
				React.DOM.form({className:"search"},
					React.DOM.input({placeholder:"Search..."}),
					React.DOM.button({
						className:"search-icon mdi mdi-magnify",
						type: "submit",
						tabIndex: -1
					})),
				a({className:"button userinfo", id:"userinfo", href:"/user:info"}, Global.User),
				React.createElement(HeaderMenu, {
					items: [
						{key:"0", href: "#", caption: "New Page"},
						{key:"1", href: "#", caption: "Company"},
						{key:"2", href: "/index:recent-changes", caption: "Recent Changes"},
						{key:"3", href: "/system/auth/logout", caption: "Logout"}
					]
				})
			);
		}
	});

	var Content = React.createClass({
		displayName: "Content",
		render: function(){
			return React.DOM.div({
				id: "content"
			},
				React.createElement(KB.Lineup.View, this.props)
			);
		}
	});

	var Site = React.createClass({
		displayName: "Site",
		render: function(){
			return React.DOM.div({},
				React.createElement(Header,  this.props),
				React.createElement(Content, this.props)
			);
		}
	});

	return Site;
})();
