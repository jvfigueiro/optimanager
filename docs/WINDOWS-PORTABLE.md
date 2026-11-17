# Guia de Execução Portátil no Windows (`optimanager.exe`)

O OptiManager pode ser executado nativamente no Windows como um executável único e portátil (`optimanager.exe`), sem necessidade de Docker, banco de dados externo ou instalação de dependências.

---

## 1. Como Executar

1. Crie uma pasta dedicada no computador principal da loja (por exemplo: `C:\OptiManager`).
2. Copie o executável `optimanager.exe` (gerado como `optimanager-windows-amd64.exe`) para dentro dessa pasta.
3. Execute o `optimanager.exe` (duplo clique ou pelo terminal).
4. Na primeira execução, o sistema criará automaticamente a subpasta `data\` ao lado do executável contendo:
   - `data\optimanager.db` (banco de dados SQLite)
   - `data\backups\` (backups compactados `.zip`)
   - `data\uploads\` (arquivos enviados)
5. Acesse no navegador deste computador: **`http://localhost:8585`**.

> **Dica (Variáveis de Ambiente Opcionais):**
> - `PORT`: altera a porta padrão `8585`.
> - `DB_PATH`: define um caminho customizado para o arquivo `.db`.
> - `JWT_SECRET`: recomenda-se definir uma chave fixa de pelo menos 32 caracteres para que os logins permaneçam válidos após reiniciar o servidor (em ambiente local sem `JWT_SECRET`, uma chave segura é gerada ou persistida conforme configuração).

---

## 2. Acesso pela Rede Local & Firewall do Windows

Para que outros computadores, tablets ou celulares na mesma rede Wi-Fi/cabo acessem o sistema pelo endereço **`http://<IP-DO-PC>:8585`**:

1. Descubra o IP local do computador servidor executando `ipconfig` no Prompt de Comando (ex.: `192.168.1.50`).
2. Libere a porta **`8585`** no Firewall do Windows executando o **PowerShell como Administrador**:

```powershell
New-NetFirewallRule -DisplayName "OptiManager (Porta 8585)" -Direction Inbound -Protocol TCP -LocalPort 8585 -Action Allow
```

3. Nos demais dispositivos da loja, acesse **`http://192.168.1.50:8585`** pelo navegador.

---

## 3. Executar Automaticamente na Inicialização (Serviço do Windows)

Para evitar que o OptiManager dependa de uma janela de terminal aberta, configure-o para iniciar automaticamente com o Windows:

### Opção A — Via NSSM (Recomendado como Serviço do Windows)
1. Baixe o **NSSM** (Non-Sucking Service Manager).
2. No Prompt de Comando como Administrador, execute:
   ```cmd
   nssm install OptiManager "C:\OptiManager\optimanager.exe"
   nssm set OptiManager AppDirectory "C:\OptiManager"
   nssm set OptiManager Start SERVICE_AUTO_START
   nssm start OptiManager
   ```

### Opção B — Via Agendador de Tarefas (`taskschd.msc`)
1. Abra o **Agendador de Tarefas** do Windows e clique em **Criar Tarefa...**.
2. Em **Geral**, marque **"Executar estando o usuário conectado ou não"** e **"Executar com privilégios mais altos"**.
3. Em **Disparadores**, crie um novo disparador **"Ao inicializar"**.
4. Em **Ações**, selecione **"Iniciar um programa"**, aponte para `C:\OptiManager\optimanager.exe` e preencha **Iniciar em (opcional)** com `C:\OptiManager`.

---

## 4. AVISO CRÍTICO: Nunca compartilhe o `.db` via pasta de rede (SMB)

> **NUNCA** coloque `optimanager.db` em uma pasta compartilhada do Windows (SMB / `\\SERVIDOR\Compartilhamento`) para ser aberto por dois executáveis `optimanager.exe` ao mesmo tempo, nem abra o arquivo `.db` simultaneamente por outros programas enquanto o servidor estiver rodando.
>
> O SQLite em modo **WAL** depende de bloqueios locais do sistema de arquivos e memória compartilhada. O acesso concorrente via SMB pode **corromper o banco de dados**.
>
> **Forma correta de uso multiusuário:** rode **apenas uma instância** do `optimanager.exe` no computador servidor e acesse de quantos dispositivos precisar via navegador em **`http://<IP-DO-SERVIDOR>:8585`**.
