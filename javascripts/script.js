async function listSponsors() {
	const table = document.getElementById("sponsors");
	if (!table || table.dataset.loaded === "true") {
		return;
	}

	table.dataset.loaded = "true";

	try {
		const response = await fetch("https://raw.githubusercontent.com/KaijuEngine/kaiju/refs/heads/master/sponsors.json");
		if (!response.ok) {
			throw new Error("Unable to load sponsors");
		}

		const sponsors = await response.json();
		sponsors.sort((a, b) => b.Support - a.Support);

		for (const sponsor of sponsors) {
			const row = document.createElement("tr");
			const name = document.createElement("td");
			name.textContent = sponsor.Name;
			row.appendChild(name);

			const github = document.createElement("td");
			const link = document.createElement("a");
			link.href = `https://github.com/${sponsor.GitHub}`;
			link.textContent = sponsor.GitHub;
			link.target = "_blank";
			link.rel = "noopener";
			github.appendChild(link);
			row.appendChild(github);
			table.appendChild(row);
		}
	} catch {
		table.dataset.loaded = "false";
	}
}

function processPage() {
	listSponsors();
}

if (typeof document$ !== "undefined") {
	document$.subscribe(processPage);
} else if (document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", processPage, { once: true });
} else {
	processPage();
}
