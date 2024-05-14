## Instalação

Para instalação das depêndencias internas é necessário seguir a seguinte [documentação](https://picpay.atlassian.net/wiki/spaces/Golang/pages/2682192453/Configura+o+para+reposit+rios+privados)

## Configuração

Para adicionar variáveis de configuração a sua aplicação deve:
- alterar o arquivo `config.go` dentro do pacote `config` para adicionar as variáveis na struct `Environments` e na função `LoadEnvVars`
- alterar o arquivo `.env.example` e fazer uma cópia dele para `.env`.

Dentro da função `LoadEnvVars`, para usar as variáveis é preciso fazer uso das funções da lib `viper` como `viper.GetString`, `viper.GetInt`, etc. Mais informações podem ser encontradas na [documentação oficial](https://github.com/spf13/viper#getting-values-from-viper).

Os valores das variáveis são definidos de acordo com a prioridade:

- variável de ambiente. Se existir uma variável de ambiente com o valor definido ela vai ter prioridade. Por exemplo, se existir a variável de ambiente `HTTP_PORT` definida o seu valor vai ser usado.
- arquivo `.env`. Se existir uma variável no arquivo `.env` ela será usada
- valor default. Se não existir valores definidos nos itens anteriores vai ser usado o valor default configurado no arquivo `main.go`. Exemplo: `viper.SetDefault("HTTP_PORT", "8066")`.
- caso não exista valor default definido será usado o `default value` do tipo (string é vazio, ponteiro é nulo, inteiro é 0)


## Rodando a aplicação localmente

- No seu terminal, na raiz do projeto execute o seguinte comando para instalar as depedências do projeto:

```sh
make dep
```

- No mesmo terminal, executa a sua aplicação executando o seguinte comando:

```shell
make run
```

## Rodando a aplicação localmente via Docker

- No seu terminal faça o build da aplicação
```shell
export GITHUB_TOKEN=<inclua aqui seu token>
export GITHUB_USER=<inclua aqui seu usuário>
make docker-build
```

- Lembre-se de substituir os valores no comando acima com o seu usuário e token do Github
- Ainda no terminal, execute:

```shell
make docker-run
```

## Pre-commit

O pre-commit é um gerenciador de pacotes multi-linguagem para hooks de pre-commit desenvolvido para resolver estes problemas. A partir de uma lista de hooks que se quer executar no projeto, o pre-commit gerencia a instalação e execução de qualquer hook antes de cada commit, incluindo o ambiente de execução da linguagem requerido, caso este não esteja instalado.

### Instalando o pre-commit

Usando pip:

    pip install pre-commit

Usando o homebrew (macOS):

    brew install pre-commit

### Configurando o pre-commit

As configurações encontram-se no arquivo `.pre-commit-config.yaml` e já encontram-se pre-configuradas para o projeto.

### Instalando os git hook scripts

Execute o comando:

    pre-commit install

Após a instalação vai sempre ser executado o gofmt, golangci-lint (executa linters paralelamente) e testes antes do commit.

Observação: caso encontre erro na execução do comando `goimports` basta executar:

    go install golang.org/x/tools/cmd/goimports@latest
