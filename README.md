

## 🔐 Why FckBot

**FckBot** is a frontend protection system designed to defend against automated bots. It leverages proof-of-work challenges and a variety of other techniques that make reverse engineering significantly more difficult. By adding layers of complexity and dynamic behavior, FckBot helps ensure that only real users can access and interact with your application effectively.

## 💻 Example of use in frontend
```html
<head>
    <script type="module" crossorigin src="{fckbot script location}"></script>
</head>
<body>
    <script>
        window.onload = () => {
            FckBot.setConfig({
                challengeUrl: "http://localhost:4000/getChallenge"
            });

            FckBot.fetch("https://localhost:4000/exampleEndpoint");
        }
    </script>
</body>
```

## 🌐 Example of use in backend

[Example in go](/backend/example.go)

```go run .```
or
```go build```

### [settings.json](/backend/settings.json)

```json
{
    "example": true,
    "host": "",
    "port": 4000,
    "apikey": "example",
    "challenge": {
        "difficulty": [8, 16],
        "latency": [100, 300]
    },
    "paths": {
        "get_challenge": "/getChallenge",
        "verify_challenge": "/verifyChallenge"
    }
}
```

| Key                    | Value       | Description                                                  |
| -----------------------|-------------|--------------------------------------------------------------|
| example                | boolean     | specifies whether cors and exampleEndpoint should be enabled |
| host                   | string      | if it is "" it works for 0.0.0.0                             |
| port                   | int         | the port on which the backend is to operate                  |
| apikey                 | string      | api key to be used for verifyChallenge                       |
| challenge.difficulty   | [int, int]  | determines the difficulty of the challenge                   |
| challenge.latency      | [int, int]  | specifies the latency in the challenge                       |
| paths.get_challenge    | string      | specifies the path for getChallenge                          |
| paths.verify_challenge | string      | specifies the path for verifyChallenge                       |

## 🔧 Technologies used

### Perfect frontend
- ⚡**Vite + Typescript**

### Very fast backend
- 🚀 **Go lang + fiber**
