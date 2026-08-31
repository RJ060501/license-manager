from src.config import load_settings
from src.collectors.entra import EntraClient


def main():
    settings = load_settings()

    entra_client = EntraClient(settings)

    skus = entra_client.get_subscribed_skus()

    print(f"Retrieved {len(skus)} subscribed SKUs.\n")

    for sku in skus:
        print(
            f"{sku.get('skuPartNumber')} | "
            f"SKU ID: {sku.get('skuId')} | "
            f"Consumed: {sku.get('consumedUnits')}"
        )


if __name__ == "__main__":
    main()