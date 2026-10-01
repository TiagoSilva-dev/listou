import { z } from "zod";
import { CuratedMerchant, CuratedProduct } from "@listou/types";
import { CuratedAdmin } from "@/components/curated-admin";
import { serverApi } from "@/lib/server-api";

export const metadata = { title: "Produtos curados · Admin" };

export default async function CuratedProductsPage() {
  const { products, merchants } = await serverApi(
    "/admin/curated-products",
    z.object({ products: z.array(CuratedProduct), merchants: z.array(CuratedMerchant) }),
    { auth: true },
  );
  return <CuratedAdmin initialProducts={products} merchants={merchants} />;
}
