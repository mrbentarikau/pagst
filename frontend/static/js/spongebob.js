lastLoc = window.location.pathname;
lastHash = window.location.hash;
$(function () {

	$("#loading-overlay").addClass("hidden");


	if (visibleURL) {
		console.log("Should navigate to", visibleURL);
		window.history.replaceState("", "", visibleURL);
	}

	addListeners();
	initPlugins(false);

	window.onpopstate = function (evt, a) {
		var shouldNav;
		console.log(window.location.pathname);
		console.log(lastLoc);
		if (window.location.pathname !== lastLoc) {
			shouldNav = true;
		} else {
			shouldNav = false;
		}

		console.log("Popped state", shouldNav, evt, evt.path);
		if (shouldNav) {
			navigate(window.location.pathname, "GET", null, false)
		}
		// Handle the back (or forward) buttons here
		// Will NOT handle refresh, use onbeforeunload for this.
	};

	if (window.location.hash) {
		navigateToAnchor(window.location.hash);
	}

	updateSelectedMenuItem(window.location.pathname);


	// Update all dropdowns
	// $(".btn-group .dropdown-menu").dropdownUpdate();
})

var currentlyLoading = false;
function navigate(url, method, data, updateHistory, maintainScroll, alertsOnly, cb) {
	if (currentlyLoading) { return; }
	closeSidebar();

	var scrollBeforeNav = document.documentElement.scrollTop;

	$("#loading-overlay").removeClass("hidden");
	// $("#main-content").html('<div class="loader">Loading...</div>');

	currentlyLoading = true;
	var evt = new CustomEvent('customnavigate', { url: url });
	window.dispatchEvent(evt);

	if (url[0] !== "/") {
		url = window.location.pathname + url;
	}

	console.log("Navigating to " + url);
	var shownURL = url;
	// Add the partial param
	var index = url.indexOf("?")
	if (index !== -1) {
		url += "&partial=1"
	} else {
		url += "?partial=1"
	}

	if (alertsOnly) {
		url += "&alertsonly=1"
	}

	PNotify.removeAll();

	updateSelectedMenuItem(url);

	var req = new XMLHttpRequest();
	req.addEventListener("load", function () {
		currentlyLoading = false;
		if (this.status !== 200 && this.status !== 400) {
			window.location.href = '/';
			return;
		} else if (this.status === 400) {
			alertsOnly = true;
		}

		if (updateHistory) {
			window.history.pushState("", "", shownURL);
		}
		lastLoc = shownURL;
		lastHash = window.location.hash;

		if (alertsOnly) {
			showAlerts(this.responseText)
			$("#loading-overlay").addClass("hidden");
			if (cb)
				cb();
			return
		}

		$("#main-content").html(this.responseText);

		initPlugins(true);
		$(document.body).trigger('ready');

		if (typeof ga !== 'undefined') {
			ga('send', 'pageview', window.location.pathname);
			console.log("Sent pageview")
		}

		if (cb)
			cb();

		if (maintainScroll)
			document.documentElement.scrollTop = scrollBeforeNav;

		$("#loading-overlay").addClass("hidden");
	});

	req.addEventListener("error", function () {
		window.location.href = '/';
		currentlyLoading = false;
	});

	req.open(method, url);
	req.setRequestHeader('Cache-Control', 'no-cache');

	if (data) {
		req.setRequestHeader("content-type", "application/x-www-form-urlencoded");
		req.send(data);
	} else {
		req.send();
	}
}

function showAlerts(alertsJson) {
	var alerts = JSON.parse(alertsJson);
	if (!alerts) return;

	const stack_bar_top = { "dir1": "down", "dir2": "right", "push": "top", "spacing1": 0, "spacing2": 0 };

	for (var i = 0; i < alerts.length; i++) {
		var alert = alerts[i];

		var notice;
		if (alert.Style === "success") {
			notice = new PNotify({
				title: alert.Message,
				text: "(Click to dismiss)",
				type: 'success',
				addclass: 'stack-bar-top click-2-close',
				stack: stack_bar_top,
				width: "100%",
				delay: 2000,
				buttons: {
					closer: false,
					sticker: false
				}
			});
		} else if (alert.Style === "danger") {
			notice = new PNotify({
				title: alert.Message,
				text: "Read the docs and contact support if you don't know what went wrong.\n(Click to dismiss)",
				type: 'error',
				addclass: 'stack-bar-top click-2-close',
				stack: stack_bar_top,
				width: "100%",
				hide: false,
				buttons: {
					closer: false,
					sticker: false
				}
			});
		} else if (alert.Style === "warning-cc-limit") {
			notice = new PNotify({
				title: alert.Message,
				text: "This warning does not affect saves.\n(Click to dismiss)",
				type: 'warning-cc-limit',
				addclass: 'stack-bar-top click-2-close',
				stack: stack_bar_top,
				width: "100%",
				hide: false,
				buttons: {
					closer: false,
					sticker: false
				}
			});
		} else if (alert.Style === "warning") {
			notice = new PNotify({
				title: alert.Message,
				text: "Read the docs and contact support if you don't know what went wrong.\n(Click to dismiss)",
				type: 'warning',
				addclass: 'stack-bar-top click-2-close',
				stack: stack_bar_top,
				width: "100%",
				hide: false,
				buttons: {
					closer: false,
					sticker: false
				}
			});
		} else {
			continue;
		}

		(function () {
			var noticeCop = notice;
			noticeCop.get().click(function () {
				noticeCop.remove();
			});
		})()
	}
}

function closeSidebar() {
	document.documentElement.classList.remove("sidebar-left-opened");

	$(window).trigger("sidebar-left-opened", {
		added: false,
		removed: true
	});
}

// Automatically marks the the menu entry corresponding with our active page as active
function updateSelectedMenuItem(pathname) {
	// Collapse all nav parents first
	var navParents = document.querySelectorAll("#menu .nav-parent");
	for (var i = 0; i < navParents.length; i++) {
		navParents[i].classList.remove("nav-expanded", "nav-active");
	}

	// Then update the nav links
	var navLinks = document.querySelectorAll("#menu .nav-link")

	var bestMatch = -1;
	var bestMatchLength = 0;
	for (var i = 0; i < navLinks.length; i++) {
		var href = navLinks[i].attributes.getNamedItem("href").value;
		if (pathname.indexOf(href) !== -1) {
			if (href.length > bestMatchLength) {
				bestMatch = i;
				bestMatchLength = href.length
			}
		}

		navLinks[i].parentElement.classList.remove("nav-active");
	}

	if (bestMatch !== -1) {
		var collapseParent = navLinks[bestMatch].parentElement.parentElement.parentElement;
		if (collapseParent.classList.contains("nav-parent")) {
			collapseParent.classList.add("nav-expanded", "nav-active");
		}

		navLinks[bestMatch].parentElement.classList.add("nav-active");
	}
}

function addAlert(kind, msg, id) {
	const alert = $(`<div/>`);
	if (id !== undefined) alert.prop("id", id)
	alert.addClass("row").append(
		$("<div/>").addClass("col-lg-12").append(
			$("<div/>").addClass("alert alert-" + kind).text(msg)
		)
	).appendTo("#alerts");
}

function addAlertPAGST(kind, msg, id) {
	var style_alert = {
		margin: "0px",
		border: "none"
	}
	var style_text = {
		color: "#abb4be",
		fontFamily: "'Open Sans',Arial, sans-serif",
		fontWeight: 400,
		fontStyle: "italic"
	}
	const alert = $(`<div/>`);
	if (id !== undefined) alert.prop("id", id)
	alert.addClass("pagst").append(
			$("<div/>").addClass("alert alertPAGST-" + kind).css(style_alert).append(
				$('<i/>').attr("title", msg).addClass("fas fa-exclamation-circle").css("color","#aa8900").text(" ").append(
					$("<a />").attr("href","https://discordstatus.com").attr("target","_blank").addClass("pagst").append(
						$("<span />").css(style_text).text(msg)
					)
				)
			)
	).appendTo("#alerts_pagst");
}

function clearAlerts() {
	$("#alerts").empty();
	$("#alerts_pagst").empty();
}

function addListeners() {
	////////////////////////////////////////
	// Async partial page loading handling
	///////////////////////////////////////

	formSubmissionEvents();

	$(document).on("click", '[data-partial-load]', function (event) {
		console.log("Clicked the link");
		event.preventDefault();

		if (currentlyLoading) { return; }

		var link = $(this);

		var url = link.attr("href");
		navigate(url, "GET", null, true);
	});

	$(document).on("click", '[data-toggle="popover"]', function (evt) {
		$('[data-toggle="popover"]').each(function (i, elem) {
			// console.log(elem, elem == evt.target);
			if (evt.currentTarget == elem) {
				return;
			}
			$(elem).popover('hide');
		})
	});

	$(document).on("click", 'a[href^="#"]', function (e) {
		//e.preventDefault();

		navigateToAnchor($.attr(this, "href"));
	})


	$(document).on('click', '.btn-add', function (e) {
		e.preventDefault();

		var currentEntry = $(this).parent().parent(),
			newEntry = $(currentEntry.clone()).insertAfter(currentEntry);

		newEntry.find('input, textarea').val('');
		newEntry.parent().find('.entry:not(:last-of-type) .btn-add')
			.removeClass('btn-add').addClass('btn-remove')
			.removeClass('btn-success').addClass('btn-danger')
			.html('<i class="fas fa-minus"></i>');
	}).on('click', '.btn-remove', function (e) {
		$(this).parents('.entry:first').remove();

		e.preventDefault();
		return false;
	});

	$(document).on('click', '.modal-dismiss', function (e) {
		e.preventDefault();
		$.magnificPopup.close();
	});

	$(window).on("sidebar-left-toggle", function (evt, data) {
		window.localStorage.setItem("sidebar_collapsed", !data.removed);
		document.cookie = "sidebar_collapsed=" + data.added + "; max-age=3153600000; path=/"
	})
}

// Initializes plugins such as multiselect, we have to do this on the new elements each time we load a partial page
function initPlugins(partial) {
	var selectorPrefix = "";
	if (partial) {
		selectorPrefix = "#main-content ";
	}

	$(selectorPrefix + '[data-toggle="popover"]').popover()
	$(selectorPrefix + '[data-toggle="tooltip"]').tooltip();

	$('.entry:not(:last-of-type) .btn-add')
		.removeClass('btn-add').addClass('btn-remove')
		.removeClass('btn-success').addClass('btn-danger')
		.html('<i class="fas fa-minus"></i>');

	// The uitlity that checks wether the bot has permissions to send messages in the selected channel
	channelRequirepermsDropdown(selectorPrefix);
	yagInitSelect2(selectorPrefix)
	yagInitMultiSelect(selectorPrefix)
	yagInitAutosize(selectorPrefix);
	yagInitUnsavedForms(selectorPrefix)
	// initializeMultiselect(selectorPrefix);

	$(selectorPrefix + '.modal-basic').magnificPopup({
		type: 'inline',
		preloader: false,
		modal: true
	});
}

var discordPermissions = {
	read: {
		name: "Read Messages",
		perm: BigInt(0x400),
	},
	send: {
		name: "Send Messages",
		perm: BigInt(0x800),
	},
	embed: {
		name: "Embed Links",
		perm: BigInt(0x4000),
	},
}
var cachedChannelPerms = {};
function channelRequirepermsDropdown(prefix) {
	var dropdowns = $(prefix + "select[data-requireperms-send]");
	dropdowns.each(function (i, rawElem) {
		trackChannelDropdown($(rawElem), [discordPermissions.read, discordPermissions.send]);
	});

	var dropdownsLinks = $(prefix + "select[data-requireperms-embed]");
	dropdownsLinks.each(function (i, rawElem) {
		trackChannelDropdown($(rawElem), [discordPermissions.read, discordPermissions.send, discordPermissions.embed]);
	});
}

function trackChannelDropdown(dropdown, perms) {
	var currentElem = $('<p class="form-control-static">Checking channel permissions for bot...</p>');
	dropdown.after(currentElem);

	dropdown.on("change", function () {
		check();
	})

	function check() {
		currentElem.text("Checking channel permissions for bot...");
		currentElem.removeClass("text-success", "text-danger");
		var currentSelected = dropdown.val();
		if (!currentSelected) {
			currentElem.text("");
		} else {
			validateChannelDropdown(dropdown, currentElem, currentSelected, perms);
		}
	}
	check();
}

function validateChannelDropdown(dropdown, currentElem, channel, perms) {
	// Expire after 5 seconds
	if (cachedChannelPerms[channel] && (!cachedChannelPerms[channel].lastChecked || Date.now() - cachedChannelPerms[channel].lastChecked < 5000)) {
		var obj = cachedChannelPerms[channel];
		if (obj.fetching) {
			window.setTimeout(function () {
				validateChannelDropdown(dropdown, currentElem, channel, perms);
			}, 1000)
		} else {
			check(cachedChannelPerms[channel].perms);
		}
	} else {
		cachedChannelPerms[channel] = { fetching: true };
		createRequest("GET", "/api/" + CURRENT_GUILDID + "/channelperms/" + channel, null, function () {
			console.log(this);
			cachedChannelPerms[channel].fetching = false;
			if (this.status != 200) {
				currentElem.addClass("text-danger");
				currentElem.removeClass("text-success");

				if (this.responseText) {
					var decoded = JSON.parse(this.responseText);
					if (decoded.message) {
						currentElem.text(decoded.message);
					} else {
						currentElem.text("Couldn't check permissions :(");
					}
				} else {
					currentElem.text("Couldn't check permissions :(");
				}
				cachedChannelPerms[channel] = null;
				return;
			}

			var channelPerms = BigInt(this.responseText);
			cachedChannelPerms[channel].perms = channelPerms;
			cachedChannelPerms[channel].lastChecked = Date.now();

			check(channelPerms);
		})
	}

	function check(channelPerms) {
		var missing = [];
		for (var i in perms) {
			var p = perms[i];
			if ((channelPerms & p.perm) != p.perm) {
				missing.push(p.name);
			}
		}

		// console.log(missing.join(", "));
		if (missing.length < 1) {
			// Has perms
			currentElem.removeClass("text-danger");
			currentElem.addClass("text-success");
			currentElem.text("");
		} else {
			currentElem.addClass("text-danger");
			currentElem.removeClass("text-success");

			currentElem.text("Missing " + missing.join(", "));
		}
	}
}

function initializeMultiselect(selectorPrefix) {
	// $(selectorPrefix+".multiselect").multiselect();
}

function formSubmissionEvents() {
	// forms.each(function(i, elem){
	// 	elem.onsubmit = submitform;
	// })

	function dangerButtonClick(evt) {
		var target = $(evt.target);
		if (target.prop("tagName") !== "BUTTON") {
			target = target.parents("button");
			if (!target) {
				return
			}
		}

		if (target.attr("noconfirm") !== undefined) {
			return;
		}

		if (target.attr("noconfirm")) {
			return
		}

		if (target.attr("formaction")) {
			return;
		}

		// console.log("aaaaa", evt, evt.preventDefault);
		if (!confirm("Are you sure you want to do this? \n\nHitting cancel requires page refresh!")) {
			evt.preventDefault(true);
			evt.stopPropagation();
		}
		// alert("aaa")
	}

	$(document).on("click", ".btn-danger", dangerButtonClick);
	$(document).on("click", ".delete-button", dangerButtonClick);

	function getRandomInt(min, max) {
		min = Math.ceil(min);
		max = Math.floor(max);
		return Math.floor(Math.random() * (max - min)) + min;
	}


	$(document).on("submit", '[data-async-form]', function (event) {
		// console.log("Clicked the link");
		event.preventDefault();

		var action = $(event.target).attr("action");
		if (!action) {
			action = window.location.pathname;
		}

		submitForm($(event.target), action, false);
	});

	$(document).on("click", 'button', function (event) {
		// console.log("Clicked the link");
		var target = $(event.target);

		if (target.prop("tagName") !== "BUTTON") {
			target = target.parents("button");
		}

		var alertsOnly = false;
		if (target.attr("data-async-form-alertsonly") !== undefined) {
			alertsOnly = true;
		}

		if (!target.attr("formaction")) return;

		if (target.hasClass("btn-danger") || target.hasClass("pagst-duplicate")|| target.attr("data-open-confirm") || target.hasClass("delete-button")) {
			var title = target.attr("title");
			if (title !== undefined && !target.hasClass("pagst-duplicate")) {
				if (!confirm("Deleting " + title + ". Are you sure you want to do this?")) {
					event.preventDefault(true);
					event.stopPropagation();
					return;
				}
			} else {
				if (!confirm("Are you sure you want to do this?")) {
					event.preventDefault(true);
					event.stopPropagation();
					return;
				}
			}
		}

		// Find the parent form using the parents or the form attribute
		var parentForm = target.parents("form");
		if (parentForm.length == 0) {
			if (target.attr("form")) {
				parentForm = $("#" + target.attr("form"));
				if (parentForm.length == 0) {
					return;
				}
			} else {
				return
			}
		}

		if (parentForm.attr("data-async-form") === undefined) {
			return;
		}

		event.preventDefault();
		console.log("Should submit using " + target.attr("formaction"), event, parentForm);
		submitForm(parentForm, target.attr("formaction"), alertsOnly);

	});
}

function submitForm(form, url, alertsOnly) {
	var serialized = serializeForm(form);

	if (!alertsOnly) {
		alertsOnly = form.attr("data-async-form-alertsonly") !== undefined;
	}

	// Keep the current tab selected
	var currentTab = null
	var tabElements = $(".tabs");
	if (tabElements.length > 0) {
		currentTab = $(".tabs a.active").attr("href")
	}

	navigate(url, "POST", serialized, false, true, alertsOnly, function () {
		hideUnsavedChangesPopup($(form)[0])
		if (currentTab) {
			$(".tabs a[href='" + currentTab + "']").tab("show");
		}
	});

	$.magnificPopup.close();
}

function serializeForm(form) {
	var serialized = form.serialize();

	form.find("[data-content-editable-form]").each(function (i, v) {
		var name = $(v).attr("data-content-editable-form")
		var value = encodeURIComponent($(v).text())
		serialized += "&" + name + "=" + value;
	})

	return serialized
}

function yagInitUnsavedForms(selectorPrefix) {
	let unsavedForms = $(selectorPrefix + "form")
	unsavedForms.each(function (i, rawElem) {
		trackForm(rawElem);
	});
}

function trackForm(form) {
	let savedVersion = serializeForm($(form));

	let hasUnsavedChanges = false

	$(form).change(function () {
		console.log("Form changed!");
		checkForUnsavedChanges();
	})

	var observer = new MutationObserver(function (mutationList, observer) {
		if (!document.body.contains(form)) {
			observer.disconnect();
			hideUnsavedChangesPopup(form);
			return;
		}

		// for (let mutation of mutationList) {
		// 	for (let removed of mutation.removedNodes) {
		// 		if (removed === form) {
		// 			observer.disconnect();
		// 			hideUnsavedChangesPopup(form);
		// 			return
		// 		}
		// 	}
		// }

		if (isSavingUnsavedForms)
			checkForUnsavedChanges();
	});

	observer.observe(document.body, { childList: true, subtree: true });

	function checkForUnsavedChanges() {
		let newVersion = serializeForm($(form));
		if (newVersion !== savedVersion) {
			console.log("Its different!");
			hasUnsavedChanges = true;
			showUnsavedChangesPopup(form);
		} else {
			hasUnsavedChanges = false;
			console.log("It's the same!");
			hideUnsavedChangesPopup(form);
		}
	}
}

let unsavedChangesStack = [];
let isSavingUnsavedForms = false;

function showUnsavedChangesPopup(form) {
	if (unsavedChangesStack.includes(form)) {
		return;
	}

	unsavedChangesStack.push(form)
	updateUnsavedChangesPopup();
}

function hideUnsavedChangesPopup(form) {
	if (!unsavedChangesStack.includes(form)) {
		return;
	}

	let index = unsavedChangesStack.indexOf(form);
	unsavedChangesStack.splice(index, 1);
	updateUnsavedChangesPopup(form)
}

function updateUnsavedChangesPopup() {
	var isMobile = false; //initiate as false
	// device detection
	
	if(/(android|bb\d+|meego).+mobile|avantgo|bada\/|blackberry|blazer|compal|elaine|fennec|hiptop|iemobile|ip(hone|od)|ipad|iris|kindle|Android|Silk|lge |maemo|midp|mmp|netfront|opera m(ob|in)i|palm( os)?|phone|p(ixi|re)\/|plucker|pocket|psp|series(4|6)0|symbian|treo|up\.(browser|link)|vodafone|wap|windows (ce|phone)|xda|xiino/i.test(navigator.userAgent) 
    || /1207|6310|6590|3gso|4thp|50[1-6]i|770s|802s|a wa|abac|ac(er|oo|s\-)|ai(ko|rn)|al(av|ca|co)|amoi|an(ex|ny|yw)|aptu|ar(ch|go)|as(te|us)|attw|au(di|\-m|r |s )|avan|be(ck|ll|nq)|bi(lb|rd)|bl(ac|az)|br(e|v)w|bumb|bw\-(n|u)|c55\/|capi|ccwa|cdm\-|cell|chtm|cldc|cmd\-|co(mp|nd)|craw|da(it|ll|ng)|dbte|dc\-s|devi|dica|dmob|do(c|p)o|ds(12|\-d)|el(49|ai)|em(l2|ul)|er(ic|k0)|esl8|ez([4-7]0|os|wa|ze)|fetc|fly(\-|_)|g1 u|g560|gene|gf\-5|g\-mo|go(\.w|od)|gr(ad|un)|haie|hcit|hd\-(m|p|t)|hei\-|hi(pt|ta)|hp( i|ip)|hs\-c|ht(c(\-| |_|a|g|p|s|t)|tp)|hu(aw|tc)|i\-(20|go|ma)|i230|iac( |\-|\/)|ibro|idea|ig01|ikom|im1k|inno|ipaq|iris|ja(t|v)a|jbro|jemu|jigs|kddi|keji|kgt( |\/)|klon|kpt |kwc\-|kyo(c|k)|le(no|xi)|lg( g|\/(k|l|u)|50|54|\-[a-w])|libw|lynx|m1\-w|m3ga|m50\/|ma(te|ui|xo)|mc(01|21|ca)|m\-cr|me(rc|ri)|mi(o8|oa|ts)|mmef|mo(01|02|bi|de|do|t(\-| |o|v)|zz)|mt(50|p1|v )|mwbp|mywa|n10[0-2]|n20[2-3]|n30(0|2)|n50(0|2|5)|n7(0(0|1)|10)|ne((c|m)\-|on|tf|wf|wg|wt)|nok(6|i)|nzph|o2im|op(ti|wv)|oran|owg1|p800|pan(a|d|t)|pdxg|pg(13|\-([1-8]|c))|phil|pire|pl(ay|uc)|pn\-2|po(ck|rt|se)|prox|psio|pt\-g|qa\-a|qc(07|12|21|32|60|\-[2-7]|i\-)|qtek|r380|r600|raks|rim9|ro(ve|zo)|s55\/|sa(ge|ma|mm|ms|ny|va)|sc(01|h\-|oo|p\-)|sdk\/|se(c(\-|0|1)|47|mc|nd|ri)|sgh\-|shar|sie(\-|m)|sk\-0|sl(45|id)|sm(al|ar|b3|it|t5)|so(ft|ny)|sp(01|h\-|v\-|v )|sy(01|mb)|t2(18|50)|t6(00|10|18)|ta(gt|lk)|tcl\-|tdg\-|tel(i|m)|tim\-|t\-mo|to(pl|sh)|ts(70|m\-|m3|m5)|tx\-9|up(\.b|g1|si)|utst|v400|v750|veri|vi(rg|te)|vk(40|5[0-3]|\-v)|vm40|voda|vulc|vx(52|53|60|61|70|80|81|83|85|98)|w3c(\-| )|webc|whit|wi(g |nc|nw)|wmlb|wonu|x700|yas\-|your|zeto|zte\-/i.test(navigator.userAgent.substr(0,4))) { 
    isMobile = true;
	}

	if (unsavedChangesStack.length == 0)  {
		$("#unsaved-changes-popup").attr("hidden", true)
	} else if (isMobile) {
		if (unsavedChangesStack.length == 1) {
			$("#unsaved-changes-message").text("You have unsaved changes, would you like to save them?");
			if (!isSavingUnsavedForms)
				$("#unsaved-changes-save-button").attr("hidden", false);
				hideReorderedRolesPopup();

		} else {
			$("#unsaved-changes-message").text("You have unsaved changes on multiple forms, save them all?");
			if (!isSavingUnsavedForms)
				$("#unsaved-changes-save-button").attr("hidden", false);
				hideReorderedRolesPopup();

		}

		$("#unsaved-changes-popup").attr("hidden", false)
		hideReorderedRolesPopup();
	}
}

function saveUnsavedChanges() {
	if (unsavedChangesStack.length == 1) {
		let form = unsavedChangesStack[0];
		var action = $(form).attr("action");
		if (!action) {
			action = window.location.pathname;
		}

		submitForm($(form), action, false);
		unsavedChangesStack = [];
		updateUnsavedChangesPopup();
	} else {
		saveNext();
	}

	function saveNext() {
		$("#unsaved-changes-save-button").attr("hidden", true);

		console.log("Saving next");
		let form = unsavedChangesStack.pop();

		let action = $(form).attr("action");
		if (!action) {
			action = window.location.pathname;
		}

		let jf = $(form)
		let serialized = serializeForm(jf);

		// let alertsOnly = jf.attr("data-async-form-alertsonly") !== undefined;
		// if (!alertsOnly) {
		// 	alertsOnly = 
		// }

		// Keep the current tab selected
		// let currentTab = null
		// let tabElements = $(".tabs");
		// if (tabElements.length > 0) {
		// 	currentTab = $(".tabs a.active").attr("href")
		// }

		navigate(action, "POST", serialized, false, true, true, function () {
			console.log("Doneso!");
			if (unsavedChangesStack.length > 0) {
				saveNext();
			} else {
				isSaving = false;
				updateUnsavedChangesPopup();
			}
		});

		$.magnificPopup.close();
	}
}

function navigateToAnchor(name) {
	name = name.substring(1);

	var elem = $("a[name=\"" + name + "\"]");
	if (elem.length < 1) {
		return;
	}

	$('html, body').animate({
		scrollTop: elem.offset().top - 60
	}, 500);

	var offset = elem.offset().top;
	console.log(offset)

	window.location.hash = "#" + name
}

function createRequest(method, path, data, cb) {
	var oReq = new XMLHttpRequest();
	oReq.addEventListener("load", cb);
	oReq.addEventListener("error", function () {
		window.location.href = '/';
	});
	oReq.open(method, path);

	if (data) {
		oReq.setRequestHeader("content-type", "application/json");
		oReq.send(JSON.stringify(data));
	} else {
		oReq.send();
	}
}

function toggleTheme() {
	var elem = document.documentElement;
	if (elem.classList.contains("dark")) {
		elem.classList.remove("dark");
		elem.classList.add("sidebar-light")
		document.cookie = "light_theme=true; max-age=3153600000; path=/"
	} else {
		elem.classList.add("dark");
		elem.classList.remove("sidebar-light")
		document.cookie = "light_theme=false; max-age=3153600000; path=/"
	}
}

function toggleNordTheme() {
   //var nordTheme = false
    var head  = document.getElementsByTagName('head')[0];
    var link  = document.createElement('link');
    link.id   = "cssId";
    link.rel  = 'stylesheet';
    link.type = 'text/css';
    // if (nordTheme) {
    var elem = document.documentElement;
	if (elem.classList.contains("dark")) {
		elem.classList.remove("dark");
		elem.classList.add("sidebar-light")
        link.href = '/static/css/custom_nord.css';
        document.cookie = "nord_theme=true; max-age=3153600000; path=/"
    } else {
    	elem.classList.add("dark");
		elem.classList.remove("sidebar-light")
        link.href = '/static/css/custom.css';
        document.cookie = "nord_theme=false; max-age=3153600000; path=/"
    }
    link.media = 'all';
    head.appendChild(link);
}

function loadWidget(destinationParentID, path) {
	createRequest("GET", path + "?partial=1", null, function () {
		$("#" + destinationParentID).html(this.responseText);
	})
}

function createDragnDrop(guildID) {
	Sortable.create(rolesList, {
		animation: 150,
		easing: "cubic-bezier(1, 0, 0, 1)",
		ghostClass: "sortable-ghost",
		chosenClass: "sortable-chosen",
		touchStartThreshold: 4,
		fallbackTolerance: 3,
		handle: '.role-handle',

		onEnd: function (evt) {
			if (evt.oldIndex == evt.newIndex) {
				return;
			}

			createRequest("POST", "/manage/" + guildID + "/rolecommands/drag_cmd", {"old_index": evt.oldIndex, "new_index": evt.newIndex, "id": evt.item[0].value}, null);
			showReorderedRolesPopup();
		},
	});
}

function showReorderedRolesPopup() {
	$("#reordered-message").text("All set! Reordering roles saves automatically :)");
	$("#reordered-roles-popup").attr("hidden", false);

	setTimeout(function() {
		hideReorderedRolesPopup();
	}, 4000);
}

function hideReorderedRolesPopup() {
	$("#reordered-roles-popup").attr("hidden", true);
}