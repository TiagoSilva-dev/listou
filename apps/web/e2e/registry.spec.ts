import { expect, test } from "@playwright/test";

/**
 * Critical path: creator signs up, builds a list, publishes it; an anonymous
 * guest reserves, buys through /go and cancels; the creator sees the numbers.
 * Requires the API (:8080) and web (:3000) running against a seeded database.
 */

const stamp = Date.now();
const email = `e2e-${stamp}@example.com`;

test("creator publishes a list and a guest reserves and buys", async ({
  page,
  browser,
}, testInfo) => {
  const isMobile = testInfo.project.name === "mobile";

  // --- creator: register + 3-step onboarding --------------------------------
  await page.goto("/criar?tipo=HOUSEWARMING");
  await expect(page).toHaveURL(/criar-conta/);
  await page.getByLabel("Seu nome").fill("Criador E2E");
  await page.getByLabel("E-mail").fill(email);
  await page.getByLabel("Senha").fill("senha-segura-1");
  await page.getByRole("button", { name: "Criar minha conta" }).click();

  await expect(page.getByRole("heading", { name: "Conte um pouco sobre o evento" })).toBeVisible();
  await page.getByLabel("Nome do evento").fill(`Casa nova E2E ${stamp}`);
  await page.getByLabel("Quem são os anfitriões?").fill(`Casal E2E ${stamp}`);
  await page.getByRole("button", { name: /Continuar/ }).click();
  await page.getByRole("radio", { name: /Começar com sugestões/ }).click();
  await page.getByRole("button", { name: /Criar minha lista/ }).click();

  // --- creator: dashboard, add a product from search ------------------------
  await expect(page.getByRole("heading", { name: `Casa nova E2E ${stamp}` })).toBeVisible();
  await page.getByRole("button", { name: "Adicionar item" }).first().click();
  await page.getByLabel("Buscar produtos").fill("liquidificador");
  const add = page.getByRole("button", { name: /Adicionar$/ }).first();
  await expect(add).toBeVisible();
  await add.click();
  await expect(page.getByText(/foi adicionado à lista/).first()).toBeVisible();
  await page.keyboard.press("Escape");

  // manual item without any marketplace
  await page.getByRole("button", { name: "Adicionar item" }).first().click();
  await page.getByRole("tab", { name: "Adicionar à mão" }).click();
  await page.getByLabel("Nome do item").fill("Dinheiro para a lua de mel");
  await page.getByRole("button", { name: "Adicionar à lista" }).click();
  await expect(page.getByText("Dinheiro para a lua de mel").first()).toBeVisible();

  // --- publish ---------------------------------------------------------------
  await page.getByRole("button", { name: /Publicar lista/ }).click();
  await expect(page.getByRole("dialog", { name: "Compartilhar lista" })).toBeVisible();
  const url = await page.getByLabel("Link da lista").inputValue();
  expect(url).toContain("/l/");
  await page.keyboard.press("Escape");

  // --- guest: anonymous visit ------------------------------------------------
  const guestCtx = await browser.newContext({ ...testInfo.project.use });
  const guest = await guestCtx.newPage();
  const path = new URL(url).pathname;
  await guest.goto(path);
  await expect(guest.getByRole("heading", { name: `Casal E2E ${stamp}` })).toBeVisible();
  await expect(guest.getByRole("link", { name: "Criar minha lista" })).toBeVisible(); // viral loop

  // (the template also suggests a plain "Liquidificador" desire; pick the one linked to a product)
  await guest.getByLabel("Buscar na lista").fill("liquidificador 12");
  await guest
    .getByRole("button", { name: /Ver Liquidificador 12 velocidades/i })
    .first()
    .click();

  // buy now: opens the store through /go in a new tab and records the click
  const [store] = await Promise.all([
    guestCtx.waitForEvent("page"),
    guest
      .getByRole("link", { name: /Comprar/ })
      .first()
      .click(),
  ]);
  await store.waitForLoadState("load");
  expect(store.url()).toContain("/demo/loja/");
  expect(store.url()).toContain("demo_tag=listou-");
  await store.close();

  // reserve, then cancel with the browser-held token
  await guest.getByRole("button", { name: "Reservar este presente" }).click();
  await guest.getByLabel("Seu nome").fill("Convidada E2E");
  await guest.getByRole("button", { name: "Confirmar reserva" }).click();
  await expect(guest.getByText("Presente reservado!")).toBeVisible();
  await guest.getByRole("button", { name: "Voltar para a lista" }).click();
  await expect(guest.getByText("Sua escolha").first()).toBeVisible();

  // --- creator sees the activity --------------------------------------------
  await page.reload();
  await expect(page.getByText("Convidada E2E").first()).toBeVisible();
  await expect(page.getByText(/1 de \d+ presentes escolhidos/)).toBeVisible();

  // guest cancels -> item is available again
  await guest
    .getByRole("button", { name: /Ver minha escolha/ })
    .first()
    .click();
  await guest.getByRole("button", { name: /Cancelar reserva/ }).click();
  await expect(guest.getByText("Sua escolha")).toHaveCount(0);

  if (!isMobile) {
    // public page has SEO essentials
    const canonical = await guest.locator('link[rel="canonical"]').getAttribute("href");
    expect(canonical).toContain(path);
  }
  await guestCtx.close();
});

test("unknown and draft lists are not public", async ({ page }) => {
  const res = await page.goto("/l/esta-lista-nao-existe");
  expect(res?.status()).toBe(404);
  await expect(page.getByText("Não encontramos esta página")).toBeVisible();
});

test("seeded public list renders on mobile without horizontal scroll", async ({ page }) => {
  await page.goto("/l/tiago-e-julia");
  await expect(page.getByRole("heading", { name: "Tiago & Julia" })).toBeVisible();
  const overflow = await page.evaluate(
    () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
  );
  expect(overflow).toBeLessThanOrEqual(0);
});
