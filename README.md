# Discord Music Bot

Base em Go e Docker para um bot de Discord focado em musica.

Este scaffold usa uma camada de fila e comandos de texto com `discordgo`. A integracao de audio foi deixada agnostica de fonte por seguranca e para evitar dependencias de extracao nao autorizada do YouTube Music. Voce pode plugar depois uma fonte autorizada ou streams diretos que voce controle.

## Recursos incluidos

- Estrutura em Go com `cmd/` e `internal/`
- Configuracao via variaveis de ambiente
- Comandos basicos: `play`, `queue`, `pause`, `resume`, `skip`, `leave` e `ping`
- Dockerfile e Docker Compose para build e execucao
- Base pronta para evoluir para um adaptador de audio autorizado

## Requisitos

- Go 1.22 ou superior
- Docker, se quiser rodar em container
- Um bot do Discord com token valido
- Privileged intent de mensagem habilitado no portal do Discord, se usar comandos por prefixo

## Configuracao

1. Exporte as variaveis de ambiente:

```bash
export DISCORD_TOKEN="seu-token"
export COMMAND_PREFIX="!"
export BOT_NAME="discord-music-bot"
```

2. Rode localmente:

```bash
go mod tidy
go run ./cmd/bot
```

3. Gere a imagem Docker ou suba com Compose:

```bash
docker build -t discord-music-bot .
docker compose up --build
```

## Comandos

- `!play <consulta>` adiciona uma faixa a fila
- `!queue` mostra a fila atual
- `!pause` pausa a execucao da fila
- `!resume` retoma a execucao da fila
- `!skip` pula para a proxima faixa
- `!leave` limpa o estado do servidor
- `!ping` responde com pong

## Observacao sobre YouTube Music

A base nao implementa extracao, download ou bypass de protecao do YouTube Music. Se voce precisar de reproducao real, conecte um provedor autorizado ou uma fonte de audio que voce controle.
