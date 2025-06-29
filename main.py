import asyncio

from amcp_pylib.core import ClientAsync


async def main():
    client = ClientAsync()
    try:
        await client.connect()
    except Exception as e:
        print(f"Failed to connect to the server: {e}")
        return
    
    print("Connected to the server")
    try:
        response = await client.send("VERSION\r\n".encode("utf-8"))
        print("Server version:", response)
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        print("Exiting...")

if __name__ == "__main__":
    asyncio.run(main())