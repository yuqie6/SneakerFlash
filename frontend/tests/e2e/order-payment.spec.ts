import { expect, test } from "@playwright/test"

test("order detail can complete payment flow", async ({ page }) => {
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

  let paid = false

  await page.route("**/api/v1/orders/9", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          order: {
            id: 9,
            order_num: "ORD-001",
            product_id: 1,
            status: paid ? 1 : 0,
            created_at: "2024-01-01T00:00:00Z",
          },
          payment: {
            payment_id: "PAY-001",
            status: paid ? "paid" : "pending",
            amount_cents: 129900,
          },
        },
      }),
    })
  })

  await page.route("**/api/v1/product/1", async (route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {
          id: 1,
          name: "AJ 1 Chicago",
          stock: 8,
          price: 1299,
          image: "",
        },
      }),
    })
  })

  await page.route("**/api/v1/coupons/mine**", async (route) => {
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
          page_size: 100,
        },
      }),
    })
  })

  await page.route("**/api/v1/payment/callback", async (route) => {
    paid = true
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        code: 200,
        msg: "ok",
        data: {},
      }),
    })
  })

  await page.goto("/login?redirect=/orders/9")
  await page.getByPlaceholder("输入用户名").fill("alice")
  await page.getByPlaceholder("输入密码").fill("password-123")
  await page.getByRole("button", { name: "登录" }).click()
  await page.getByRole("button", { name: "确认支付" }).click()

  await expect(page.getByText("已支付").first()).toBeVisible()
})
