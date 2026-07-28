async function listSponsors() {
	const box = document.getElementById("sponsors");
	if (!box) {
		return;
	}

	box.querySelectorAll("[data-sponsor-row]").forEach(row => row.remove());

	try {
		const list = await fetch("https://raw.githubusercontent.com/KaijuEngine/kaiju/refs/heads/master/sponsors.json");
		if (!list.ok) {
			throw new Error("Unable to load sponsors");
		}

		const sponsors = await list.json();
		sponsors.sort((a, b) => b.Support - a.Support);

		for (const sponsor of sponsors) {
			const tr = document.createElement("tr");
			tr.dataset.sponsorRow = "true";

			const tdName = document.createElement("td");
			tdName.textContent = sponsor.Name;
			tr.appendChild(tdName);

			const tdGit = document.createElement("td");
			const a = document.createElement("a");
			a.href = `${sponsor.GitHub}`;
			a.textContent = sponsor.GitHub;
			a.target = "_blank";
			a.rel = "noopener";
			tdGit.appendChild(a);
			tr.appendChild(tdGit);

			box.appendChild(tr);
		}
	} catch {
		const tr = document.createElement("tr");
		tr.dataset.sponsorRow = "true";
		const td = document.createElement("td");
		td.colSpan = 2;
		td.textContent = "Sponsors could not be loaded right now.";
		tr.appendChild(td);
		box.appendChild(tr);
	}
}

function formatPostDate(value) {
	if (!value) {
		return "";
	}

	const date = new Date(`${value}T00:00:00`);
	if (Number.isNaN(date.getTime())) {
		return value;
	}

	return new Intl.DateTimeFormat(undefined, {
		month: "short",
		day: "numeric",
		year: "numeric"
	}).format(date);
}

function createNewsCard(post, newsUrl) {
	const article = document.createElement("a");
	article.className = "kl-card kl-blog-card";
	article.href = new URL(post.url || "../blog/", newsUrl).pathname;
	article.setAttribute("aria-label", `Read ${post.title || "Kaiju update"}`);

	const category = document.createElement("span");
	category.textContent = post.category || "Update";
	article.appendChild(category);

	const title = document.createElement("h3");
	title.textContent = post.title || "Kaiju update";
	article.appendChild(title);

	const time = document.createElement("time");
	time.dateTime = post.date || "";
	time.textContent = formatPostDate(post.date);
	article.appendChild(time);

	const description = document.createElement("p");
	description.textContent = post.description || "";
	article.appendChild(description);

	const link = document.createElement("span");
	link.className = "kl-text-link";
	link.innerHTML = 'Read post <span aria-hidden="true">-&gt;</span>';
	article.appendChild(link);

	return article;
}

async function loadNews() {
	const list = document.getElementById("kaiju-news-list");
	if (!list) {
		return;
	}

	const source = list.dataset.newsSrc || "blog/posts.json";
	const newsUrl = new URL(source, document.baseURI);

	try {
		const response = await fetch(newsUrl);
		if (!response.ok) {
			throw new Error("Unable to load news");
		}

		const posts = await response.json();
		const sortedPosts = posts
			.slice()
			.sort((a, b) => String(b.date || "").localeCompare(String(a.date || "")))
			.slice(0, 3);

		list.replaceChildren(...sortedPosts.map(post => createNewsCard(post, newsUrl)));
	} catch {
		const fallback = document.createElement("p");
		fallback.className = "kl-news-loading";
		fallback.innerHTML = 'News could not be loaded right now. <a class="kl-text-link" href="blog/">Open the blog <span aria-hidden="true">-&gt;</span></a>';
		list.replaceChildren(fallback);
	}
}

function setupShowcase() {
	const video = document.getElementById("kaiju-showcase-video");
	const title = document.getElementById("kaiju-showcase-title");
	const description = document.getElementById("kaiju-showcase-description");
	const buttons = document.querySelectorAll(".kl-showcase-thumb");
	const prefersReducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

	buttons.forEach(button => {
		if (button.dataset.showcaseReady === "true") {
			return;
		}

		button.dataset.showcaseReady = "true";
		button.addEventListener("click", () => {
			buttons.forEach(item => {
				item.classList.remove("is-active");
				item.setAttribute("aria-pressed", "false");
			});

			button.classList.add("is-active");
			button.setAttribute("aria-pressed", "true");

			if (title) {
				title.textContent = button.dataset.title || "";
			}

			if (description) {
				description.textContent = button.dataset.description || "";
			}

			if (video instanceof HTMLVideoElement && button.dataset.video) {
				const source = video.querySelector("source");
				if (source) {
					source.src = button.dataset.video;
				}

				if (button.dataset.poster) {
					video.poster = button.dataset.poster;
				}

				video.load();

				if (!prefersReducedMotion) {
					video.play().catch(() => {});
				}
			}
		});
	});
}

async function processIndex() {
	if (!document.querySelector(".kaiju-landing")) {
		return;
	}

	setupShowcase();
	await Promise.all([listSponsors(), loadNews()]);
}

function onReady(callback) {
	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", callback, { once: true });
	} else {
		callback();
	}
}

if (typeof document$ !== "undefined") {
	document$.subscribe(() => {
		processIndex();
	});
} else {
	onReady(() => {
		processIndex();
	});
}
