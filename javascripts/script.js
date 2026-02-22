async function listSponsors() {
	let box = document.getElementById("sponsors");
	let list = await fetch("https://raw.githubusercontent.com/KaijuEngine/kaiju/refs/heads/master/sponsors.json");
	let sponsors = await list.json()
	sponsors.sort((a, b) => b.Support - a.Support);
	for (let i = 0; i < sponsors.length; i++) {
		const s = sponsors[i];
		const tr = document.createElement('tr');
		const tdName = document.createElement('td');
		tdName.textContent = s.Name;
		tr.appendChild(tdName);
		const tdGit = document.createElement('td');
		const a = document.createElement('a');
		a.href = `https://github.com/${s.GitHub}`;
		a.textContent = s.GitHub;
		a.target = '_blank';
		a.rel = 'noopener';
		tdGit.appendChild(a);
		tr.appendChild(tdGit);
		box.appendChild(tr);
	}
}

async function processIndex() {
	// document.querySelector(".md-sidebar").style.display = "none";
	await listSponsors();
}

if (window.location.pathname == "/") {
	processIndex();
} else {
	// document.querySelector(".md-sidebar").style.display = "";
}
