package("kb.boot", function(exports) {
	"use strict";

	depends("boot.css");

	depends("Login.js");
	depends("Session.js");

	depends("Crumbs.js");
	depends("Lineup.js");
	depends("Site.js");
	depends("Selection.js");

	var Bootstrap = React.createClass({
		componentDidMount: function() {
			var self = this;
			if (typeof Reloader !== "undefined") {
				Reloader.onchange = function() {
					self.forceUpdate();
				};
			}
		},
		componentWillUnmount: function() {
			var session = this.state.Session;
			if (session !== null) {
				session.remove(this);
			}
		},

		getInitialState: function() {
			return {
				Session: null,
				sessionError: ""
			};
		},

		sessionFinished: function(ev) {
			this.setState({
				Session: null,
				sessionError: ev.error
			});
		},
		loggedIn: function(sessionInfo) {
			var session = new kb.Session(sessionInfo);
			session.on("session-finished", this.sessionFinished, this);
			this.setState({
				Session: session
			});
		},
		render: function() {
			var session = this.state.Session;
			if (session === null) {
				return React.createElement(kb.boot.Login, {
					onSuccess: this.loggedIn,
					initialError: this.state.sessionError,
					providers: window.LoginProviders
				});
			}

			return React.createElement(Application, this.state);
		}
	});

	var Application = React.createClass({
		getInitialState: function() {
			var session = this.props.Session;
			var lineup = new kb.Lineup(session);
			var crumbs = new kb.Crumbs(lineup);
			return {
				Lineup: lineup,
				Crumbs: crumbs,
				CurrentSelection: new kb.Selection(),
				Session: session
			};
		},
		childContextTypes: {
			Lineup: React.PropTypes.object,
			Crumbs: React.PropTypes.object,
			CurrentSelection: React.PropTypes.object,
			Session: React.PropTypes.object
		},
		getChildContext: function() {
			return this.state;
		},

		keydown: function(ev) {
			ev = ev || event;

			function elementIsEditable(elem) {
				return elem && (
					((elem.nodeName === "INPUT") && (elem.type === "text")) ||
					(elem.nodeName === "TEXTAREA") ||
					(elem.contentEditable === "true")
				);
			}

			if (ev.defaultPrevented || elementIsEditable(ev.target)) {
				return;
			}
			if (ev.keyCode === 27) {
				this.state.Lineup.closeLast();
				ev.preventDefault();
				ev.stopPropagation();
			}
		},
		click: function(ev) {
			ev = ev || event;
			this.state.Lineup.handleClickLink(ev);
		},

		componentDidMount: function() {
			document.onkeydown = this.keydown;
			document.onclick = this.click;
			this.state.Crumbs.attach(this.state.Session.home);
		},
		componentWillUnmount: function() {
			if (document.onkeydown === this.keydown) {
				document.onkeydown = null;
			}
			if (document.onclick === this.click) {
				document.onclick = null;
			}
			this.state.Crumbs.detach();
		},
		render: function() {
			return React.createElement(kb.Site, this.state);
		}
	});

	var bootstrap = React.createElement(Bootstrap);
	ReactDOM.render(bootstrap, document.getElementById("boot"));
});
