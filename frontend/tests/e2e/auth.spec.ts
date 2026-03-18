import { expect, test } from "@playwright/test"

test("login redirects to protected orders page", async ({ page }) => {
  await page.route("**/api/v1/login", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          access_token: "access-token",
          refresh_token: "refresh-token",
          expires_in: 3600,
        },
      }),
    })
  })

  await page.route("**/api/v1/profile", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          id: 1,
          username: "alice",
          avatar: "",
        },
      }),
    })
  })

  await page.route("**/api/v1/orders**", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          list: [],
          total: 0,
          page: 1,
          page_size: 10,
        },
      }),
    })
  })

  await page.goto("/login?redirect=/orders")
  await page.getByPlaceholder("输入用户名").fill("alice")
  await page.getByPlaceholder("输入密码").fill("password-123")
  await page.getByRole("button", { name: "登录" }).click()

  await expect(page).toHaveURL(/\/orders$/)
  await expect(page.getByRole("heading", { name: "订单中心" })).toBeVisible()
})
