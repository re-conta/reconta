Ao criar uma conta deve ser enviado para o e-mail de cadastro um código OTP, somente após a validação deste código a conta deve ser efetivamente criada. Após 2 horas se a conta não for apagada a mesma deve ser encerrada, no caso do otp perdido o usuário poderá re-enviar o código.

Em /exportar não inclua os gráficos, porem os gráficos devem ser mostrados no pdf exportado.

Em /relatorios sob "Tudo" e "Intervalo personalizado" eu preciso do fluxo tambem, preciso de um gráfico por linha e não um ao lado do outro.

No painel de admin crie uma página de logs onde toda visita deve ser logada assim como agente, ip, navegador, sistema e as páginas que visitou e as ações que este usuário fez no site.

Transforme a pasta files/ em ansible/ crie um playbook simples mais completo usando o ansible com:

- Suporte as units systemd do go
- Proxy reverso para o vue e go usando o nginx
- Instalação de certificados letsencrypt(caso eles não existam, cuidado com o block por parte do cloudflare/letsencrypt)
- Instalação do go, nginx, bun, node(se ainda não estiverem instalados), mas fique atento o bun na vps é instalado na home do usuário nginx
- Criação dos usuários e grupos caso ainda não existam(nginx por exemplo)
