document.addEventListener('DOMContentLoaded', function () {
    // Configure marked options
    marked.setOptions({
        breaks: true,
        gfm: true,
        headerIds: false
    });

    var sendButton = document.getElementById('send-button');
    sendButton.addEventListener('click', function() {
        var input = document.getElementById("chat-input");
        var message = input.value.trim();
        if (message) {
            displayMessage(message, "You");
            input.value = "";
            fetch('/chatbot', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ question: message })
            })
            .then(response => response.json())
            .then(data => {
                displayMessage(data.response, "Bot");
            })
            .catch(error => console.error('Error:', error));
        }
    });

    // Add enter key support
    document.getElementById('chat-input').addEventListener('keypress', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendButton.click();
        }
    });
});

function displayMessage(message, sender) {
    var chatHistory = document.getElementById("chat-history");
    var messageDiv = document.createElement("div");
    messageDiv.classList.add("chat-message");
    
    if (sender === "Bot") {
        // Parse markdown for bot messages
        messageDiv.innerHTML = `<strong>${sender}:</strong> ${marked.parse(message)}`;
    } else {
        // Regular text for user messages
        messageDiv.innerHTML = `<strong>${sender}:</strong> ${message}`;
    }
    
    chatHistory.appendChild(messageDiv);
    chatHistory.scrollTop = chatHistory.scrollHeight;
}
