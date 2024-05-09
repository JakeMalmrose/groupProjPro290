using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using AutoMapper;
 
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Authorization;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
 
 
[ApiController]
[Route("[controller]")]
public class AuthController : ControllerBase
{
    private readonly ILogger<OrderController> _logger;
    private readonly IMapper _mapper;
    private readonly OrderServiceDBContext _db;
    private readonly IConfiguration _config;
 
 
    public AuthController(ILogger<OrderController> logger, OrderServiceDBContext db, IMapper mapper, IConfiguration iconfig)
    {
        _logger = logger;
        _db = db;
        _mapper = mapper;
        _config = iconfig;
    }
 
    [HttpGet]
    [Route("Test1")]
    public async Task<ActionResult<List<Order>>> Test1()
    {
        return Ok("hello from AuthController");
    }
 
   
 
    [HttpPost]
    [Route("CreateTokenMethod1")]
    public async Task<ActionResult<String>> CreateTokenMethod1(UserDTO userDTO)
    {
        //if (user.Email == "ccantera@yahoo.com" && user.Password == "pwd123")
        //User user = await _userRepository.GetByCredentials(user.Email, user.Password);
 
        User? user = await _db.Users.Include(u => u.Orders).Where(u => u.Email == userDTO.Email && u.Password == userDTO.Password).FirstOrDefaultAsync();
 
        if (user != null)
        {
            var authClaims = new List<Claim> {
                new(ClaimTypes.Name, user.Username),
                new(ClaimTypes.Email, user.Email),
                new Claim("UserGuid", user.UserGuid.ToString())
                };
            var authSigningKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_config["Jwt:Key"]));
 
            var token = new JwtSecurityToken(
                issuer: _config["Jwt:Issuer"],
                audience: _config["Jwt:Audience"],
                expires: DateTime.Now.AddHours(3),
                claims: authClaims,
                signingCredentials: new SigningCredentials(authSigningKey, SecurityAlgorithms.HmacSha256)
            );
 
            string finalToken = new JwtSecurityTokenHandler().WriteToken(token);
 
            return Ok(finalToken);
        }
 
        return Unauthorized();
    }
 
 
    // another great example from that InfoWorld link, probably very similair to above, but wanted two good examples
    // this one requires an even larger signing key
    // works great, i like the output inspected in jwt.io more for this one, and it has a stronger algorithm than the one above.
    // both produced a valid key, which can then call the /security/getmessage endpoint for proof
 
    [HttpPost]
    [Route("CreateTokenMethod2")]
    public async Task<IResult> CreateTokenMethod2(UserDTO userDTO)
    {
        //if (user.Email == "ccantera@yahoo.com" && user.Password == "pwd123")
        //User user = await _userRepository.GetByCredentials(user.Email, user.Password);
 
        User? user = await _db.Users.Include(u => u.Orders).Where(u => u.Email == userDTO.Email && u.Password == userDTO.Password).FirstOrDefaultAsync();
 
        if (user != null)
        {
            var issuer = _config["Jwt:Issuer"];
            var audience = _config["Jwt:Audience"];
            var key = Encoding.ASCII.GetBytes(_config["Jwt:Key"]);
            var tokenDescriptor = new SecurityTokenDescriptor
            {
                Subject = new ClaimsIdentity(new[]
                {
                new Claim("Id", Guid.NewGuid().ToString()),
                new Claim(JwtRegisteredClaimNames.Sub, user.Username),
                new Claim(JwtRegisteredClaimNames.Email, user.Username),
                new Claim(JwtRegisteredClaimNames.Jti,
                Guid.NewGuid().ToString())
            }),
                Expires = DateTime.UtcNow.AddMinutes(5),
                Issuer = issuer,
                Audience = audience,
                SigningCredentials = new SigningCredentials
                (new SymmetricSecurityKey(key),
                SecurityAlgorithms.HmacSha512Signature)
            };
            var tokenHandler = new JwtSecurityTokenHandler();
            var token = tokenHandler.CreateToken(tokenDescriptor);
            var jwtToken = tokenHandler.WriteToken(token);
            var stringToken = tokenHandler.WriteToken(token);
            return Results.Ok(stringToken);
        }
 
        return Results.Unauthorized();
    }
 
}
