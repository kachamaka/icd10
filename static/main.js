document.addEventListener("DOMContentLoaded", () => {
    const searchButton = document.getElementById("searchButton");

    searchInput.addEventListener("keydown", (event) => {
        if (event.key === "Enter") {
            event.preventDefault();  // Prevent form submission if in a form
            search();
        }
    });

    searchButton.addEventListener("click", () => {
        search();
    });
});

function search() {
    const searchInput = document.getElementById("searchInput");
    const resultsDiv = document.getElementById("results");

    const query = searchInput.value.trim();
    if (!query) return;

    resultsDiv.innerHTML = "";

    let req = {
        "query": query
    }  

    fetch("http://localhost:8080/search", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(req),
    })
    .then((response) => {
        if (!response.ok) {
            throw new Error("Failed to fetch results");
        }
        return response.json();
    })
    .then((result) => {
        console.log(result);
        icd10codes = result.icd10codes;
        console.log(icd10codes);
        if (!icd10codes) {
            resultsDiv.innerHTML = "<p>No results found</p>";
            return;
        }
        if (icd10codes.length > 0) {
            const firstResult = icd10codes[0];
            const firstResultDiv = document.createElement("div");
            firstResultDiv.classList.add("result-item");
            firstResultDiv.style.backgroundColor = "#e0f7fa";
            firstResultDiv.style.border = "2px solid #00796b";

            firstResultDiv.innerHTML = `
                <strong>ICD-10 Code: ${firstResult.icd10code}</strong>
                Score: ${firstResult.score.toFixed(2)}
                Title: ${firstResult.title}
                Chapter: ${firstResult.chapterCode}
                Block: ${firstResult.blockCode}
                Category: ${firstResult.categoryCode}
                Type: ${firstResult.type}
                Symptoms: ${firstResult.symptoms}
            `;
            resultsDiv.appendChild(firstResultDiv);

            // If there are additional results, display them below
            if (icd10codes.length > 1) {
                const additionalResultsDiv = document.createElement("div");
                additionalResultsDiv.classList.add("result-item");
                additionalResultsDiv.innerHTML = "<strong>Additional Results:</strong>";
                resultsDiv.appendChild(additionalResultsDiv);

                // Display remaining results as additional results
                icd10codes.slice(1).forEach((result) => {
                    const resultDiv = document.createElement("div");
                    resultDiv.classList.add("result-item");
                    resultDiv.innerHTML = `
                        ICD-10 Code: ${result.icd10code}
                        Score: ${result.score.toFixed(2)}
                        Title: ${result.title}
                        Chapter: ${result.chapterCode}
                        Block: ${result.blockCode}
                        Category: ${result.categoryCode}
                        Type: ${result.type}
                        Symptoms: ${result.symptoms}
                    `;
                    additionalResultsDiv.appendChild(resultDiv);
                });
            }
        } else {
            resultsDiv.innerHTML = "<p>No results found</p>";
        }
    })
    .catch((error) => {
        resultsDiv.innerHTML = `<p>Error: ${error.message}</p>`;
    });
}
