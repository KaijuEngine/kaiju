async function listSupporters() {
	let box = document.getElementById("supporters");
	let list = await fetch("https://raw.githubusercontent.com/KaijuEngine/kaiju/refs/heads/master/sponsors.json");
	let supporters = await list.json()
	supporters.sort((a, b) => b.Support - a.Support);
	for (let i = 0; i < supporters.length; i++) {
		const s = supporters[i];
		const tr = document.createElement('tr');
		const tdName = document.createElement('td');
		tdName.textContent = s.Name;
		tr.appendChild(tdName);
		const tdLevel = document.createElement('td');
		tdLevel.textContent = s.Support;
		tr.appendChild(tdLevel);
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
	await listSupporters();
}

if (window.location.pathname == "/") {
	processIndex();
} else {
	// document.querySelector(".md-sidebar").style.display = "";
}
