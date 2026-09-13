let btns = [];
let containers = document.getElementsByClassName("icons__group");
for (let i = 0; i < containers.length; i++) {
	let all = containers[i].getElementsByTagName("button");
	for (let j = 0; j < all.length; j++) {
		if (all[i] && all[i].getAttribute("aria-haspopup") == "dialog"
			&& all[i].getAttribute("role") == "option")
		{
			btns.push(all[j]);
		}
	}
}
console.log(btns);
let entries = [];
let fields = ["Code point", "Icon name"];
let delay = 0;
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
(async () => {
	for (let i = 0; i < btns.length; i++) {
		try {
			btns[i].click();
			await sleep(100);
			try {
				let sections = document.getElementsByClassName("side-nav-links__path-wrapper ng-star-inserted");
				let tmp = {};
				for (let j = 0; j < fields.length; j++) {
					for (let k = 0; k < sections.length; k++) {
						if (sections[k].textContent.indexOf(fields[j]) >= 0) {
							let txt = sections[k].getElementsByClassName("code-snippet__content")[0].textContent;
							tmp[fields[j]] = txt;
						}
					}
				}
				entries.push(tmp);
			} catch {}
		} catch {}
	}
	console.log(entries);
	console.log(JSON.stringify(entries));
})();
