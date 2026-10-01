import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { Avatar, Button, CategoryTabs, cn, glyphName, Price, Progress } from "../index";

afterEach(cleanup);

describe("ui", () => {
  it("keeps custom font-size tokens next to text colors", () => {
    expect(cn("text-display", "text-ink")).toBe("text-display text-ink");
  });

  it("renders a loading button as disabled and busy", () => {
    render(<Button loading>Salvar</Button>);
    const btn = screen.getByRole("button", { name: "Salvar" });
    expect(btn).toHaveProperty("disabled", true);
    expect(btn.getAttribute("aria-busy")).toBe("true");
  });

  it("formats prices and shows discount only when lower", () => {
    const { container } = render(<Price cents={37900} originalCents={39900} />);
    expect(container.textContent?.replace(/\s/g, " ")).toContain("R$ 379,00");
    expect(container.querySelector("s")).not.toBeNull();
  });

  it("clamps progress", () => {
    render(<Progress value={30} max={20} label="Progresso" />);
    expect(screen.getByRole("progressbar").firstElementChild?.getAttribute("style")).toContain(
      "100%",
    );
  });

  it("shows initials for couples", () => {
    render(<Avatar name="João & Maria" />);
    expect(screen.getByLabelText("João & Maria").textContent).toBe("TJ");
  });

  it("switches category tabs", () => {
    const onChange = vi.fn();
    render(
      <CategoryTabs
        tabs={[
          { id: "all", label: "Tudo" },
          { id: "k", label: "Cozinha", emoji: "pan" },
        ]}
        value="all"
        onChange={onChange}
      />,
    );
    fireEvent.click(screen.getByRole("tab", { name: /Cozinha/ }));
    expect(onChange).toHaveBeenCalledWith("k");
  });

  it("maps legacy emoji (with or without variation selector) to Listou glyphs", () => {
    expect(glyphName("🍳")).toBe("pan");
    expect(glyphName("🛏️")).toBe("bed");
    expect(glyphName("gift")).toBe("gift");
    expect(glyphName("🦄")).toBe("gift");
    expect(glyphName(null)).toBe("gift");
  });
});
