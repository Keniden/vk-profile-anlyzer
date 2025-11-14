const vkIdInput = document.getElementById("vkIdInput");
const analyzeBtn = document.getElementById("analyzeBtn");
const resultEl = document.getElementById("result");

async function analyzeProfile() {
  const vkId = vkIdInput.value.trim();
  if (!vkId) {
    alert("Введите VK ID");
    return;
  }

  resultEl.textContent = "Запуск анализа...";

  try {
    const resp = await fetch(`/profiles/${vkId}/analyze`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!resp.ok) {
      const error = await resp.json().catch(() => ({}));
      resultEl.textContent = `Ошибка: ${resp.status} ${
        error.error || ""
      }`.trim();
      return;
    }

    const data = await resp.json();
    resultEl.textContent = JSON.stringify(data, null, 2);
  } catch (e) {
    console.error(e);
    resultEl.textContent = "Ошибка сети";
  }
}

analyzeBtn.addEventListener("click", analyzeProfile);

