document.getElementById('sendButton').addEventListener('click', async () => {
    const inputValue = document.getElementById('myInput').value;
    const responseContainer = document.getElementById('responseContainer');
    const targetUrl = '/shorten';

    try {
        const response = await fetch(`${targetUrl}?url=${encodeURIComponent(inputValue)}`, {
            method: 'POST',
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const responseData = await response.text();
        responseContainer.textContent = responseData;
    } catch (error) {
        responseContainer.textContent = `Error: ${error.message}`;
        console.error('Error during POST request:', error);
    }
});