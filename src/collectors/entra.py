import logging

import requests
from azure.core.exceptions import AzureError
from azure.identity import ClientSecretCredential


logger = logging.getLogger(__name__)


class EntraClient:
    """
    Client used to authenticate to Microsoft Graph and retrieve
    Entra user and licensing information.

    The client uses app-only authentication, meaning it authenticates
    as the registered application rather than as a signed-in user.

    Required .env values:
    - TENANT_ID
    - CLIENT_ID
    - CLIENT_SECRET

    Required API permissions:
    - Microsoft Graph -> Application -> User.Read.All
    - Microsoft Graph -> Application -> LicenseAssignment.Read.All
    """

    def __init__(self, settings):
        """
        Store configuration and create the Azure credential object.

        Args:
            settings:
                Dictionary returned by the project settings/config loader.
        """
        self.settings = settings

        self.tenant_id = settings["tenant_id"]
        self.client_id = settings["client_id"]
        self.client_secret = settings["client_secret"]

        if not self.tenant_id or not self.client_id or not self.client_secret:
            raise ValueError(
                "Missing TENANT_ID, CLIENT_ID, or CLIENT_SECRET in environment."
            )

        self.credential = ClientSecretCredential(
            tenant_id=self.tenant_id,
            client_id=self.client_id,
            client_secret=self.client_secret,
        )

        self.graph_scope = "https://graph.microsoft.com/.default"
        self.base_url = "https://graph.microsoft.com/v1.0"

    def _get_access_token(self):
        """
        Acquire a short-lived bearer token for Microsoft Graph.
        """
        try:
            token = self.credential.get_token(self.graph_scope)
            return token.token

        except AzureError as exception:
            logger.error(
                "Failed to acquire Microsoft Graph access token: %s",
                exception,
            )
            raise

    def _get_headers(self):
        """
        Build the headers required for Graph API requests.
        """
        return {
            "Authorization": f"Bearer {self._get_access_token()}",
            "Content-Type": "application/json",
        }

    def get_users(self):
        """
        Retrieve Entra users and basic account information.

        Returns:
            A list of user dictionaries.
        """
        url = (
            f"{self.base_url}/users"
            "?$select=id,displayName,userPrincipalName,accountEnabled"
        )

        users = []

        while url:
            try:
                response = requests.get(
                    url,
                    headers=self._get_headers(),
                    timeout=30,
                )

                response.raise_for_status()
                data = response.json()

                users.extend(data.get("value", []))

                url = data.get("@odata.nextLink")

            except requests.RequestException as exception:
                logger.error(
                    "Failed to retrieve Entra users: %s",
                    exception,
                )
                raise

        return users
    
    def get_subscribed_skus(self):
        """
        Retrieve all Microsoft license SKUs available in the tenant.

        This allows us to map opaque SKU IDs to readable names such as
        Microsoft 365 E3.

        Returns:
            A list of subscribed SKU dictionaries.
        """

        url = f"{self.base_url}/subscribedSkus"

        try:
            response = requests.get(
                url,
                headers=self._get_headers(),
                timeout=30,
            )

            response.raise_for_status()

            data = response.json()

            return data.get("value", [])

        except requests.RequestException as exception:
            logger.error(
                "Failed to retrieve subscribed SKUs: %s",
                exception,
            )
            raise