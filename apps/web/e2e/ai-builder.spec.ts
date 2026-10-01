import { expect, test } from "@playwright/test";

/**
 * M10: the owner asks for ideas, picks some and they land on the list.
 * Skipped unless the API runs with FEATURE_FLAGS=AI_LIST_BUILDER=true.
 */
test("owner adds suggested desires to the list", async ({ page }) => {
  const flags = await page.request.get("/api/v1/flags");
  const enabled = (await flags.json()).flags?.AI_LIST_BUILDER === true;
  test.skip(!enabled, "AI_LIST_BUILDER is off");

  const email = `ai-${Date.now()}-${test.info().project.name}@example.com`;
  const headers = { origin: "http://localhost:3000" };
  const reg = await page.request.post("/api/v1/auth/register", {
    headers,
    data: { name: "Dona da Lista", email, password: "senha-segura-1" },
  });
  expect(reg.ok()).toBeTruthy();
  const created = await page.request.post("/api/v1/events", {
    headers,
    data: { type: "HOUSEWARMING", title: "Casa nova com ideias" },
  });
  const { id } = (await created.json()).event;

  await page.goto(`/dashboard/${id}`);
  await expect(page.getByText("Sua lista está vazia")).toBeVisible();
  await page.getByRole("button", { name: "Ideias" }).click();
  await page.getByLabel("Conte o que você procura").fill("banheiro");
  await page.getByRole("button", { name: "Sugerir" }).click();

  const bathroom = page.getByRole("region", { name: "Banheiro" });
  await expect(bathroom).toBeVisible();
  await bathroom.getByRole("checkbox").first().check();
  const add = page.getByRole("button", { name: /^Adicionar \d+ (item|itens) à lista$/ });
  await add.click();

  await expect(page.getByText(/adicionados?/).first()).toBeVisible();
  await expect(page.getByText("Sua lista está vazia")).toBeHidden();
});
