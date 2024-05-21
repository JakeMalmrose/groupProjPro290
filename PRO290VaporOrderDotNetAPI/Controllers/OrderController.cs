using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;
using AutoMapper;
using Microsoft.AspNetCore.Authorization;
 
 
[ApiController]
[Route("[controller]")]
public class OrderController : ControllerBase
{
    private readonly ILogger<OrderController> _logger;
    private readonly IMapper _mapper;
    private readonly OrderServiceDBContext _db;
 
 public OrderController(ILogger<OrderController> logger, OrderServiceDBContext db, IMapper mapper)
    {
        _logger = logger;
        _db = db;
        _mapper = mapper;
 
        var utcnow = DateTime.UtcNow;
        var localnow = DateTime.Now;
        string hostName = System.Net.Dns.GetHostName();
        string myIP = System.Net.Dns.GetHostEntry(hostName).AddressList[0].ToString();
 
        Console.WriteLine($"OrderController Constructor now: {localnow.ToString("MM/dd/yyyy H:mm")}, utc now: {utcnow.ToString("MM/dd/yyyy H:mm")}");
        Console.WriteLine($"OrderController Constructor, my IP Address is: {myIP}");
 
    }
    
[HttpGet]
    [Route("Test1")]
    public async Task<IActionResult> Test1()
    {
        return Ok("hello from OrderController");
    }
     // GET: api/Order
    [HttpGet]
    public async Task<ActionResult<IEnumerable<OrderDTO>>> GetOrders()
    {
        var orders = await _context.Orders.Include(o => o.OrderGames).ToListAsync();
        return _mapper.Map<List<OrderDTO>>(orders);
    }
    [HttpGet]
    [Route("GetMyIP")]
    public async Task<ActionResult<String>> GetMyIP()
    {
        var utcnow = DateTime.UtcNow;
        var localnow = DateTime.Now;
        string hostName = System.Net.Dns.GetHostName();
        string myIP = System.Net.Dns.GetHostEntry(hostName).AddressList[0].ToString();
 
        //Console.WriteLine(Environment.GetEnvironmentVariable("MY_ENV_VARIABLE01")); //coming from launchSettings.json
        //Console.WriteLine(Environment.GetEnvironmentVariable("MY_ENV_VARIABLE02")); // coming from launch.json
        //Console.WriteLine("ASPNETCORE_ENVIRONMENT: " + Environment.GetEnvironmentVariable("ASPNETCORE_ENVIRONMENT"));
 
        string s = $"now: {localnow.ToString("MM/dd/yyyy H:mm")}, utc now: {utcnow.ToString("MM/dd/yyyy H:mm")}, my IP Address is: {myIP}, my machine name is: {Environment.MachineName}\n";
        s += $"MY_ENV_VARIABLE01: {Environment.GetEnvironmentVariable("MY_ENV_VARIABLE01c")}, MY_ENV_VARIABLE02: {Environment.GetEnvironmentVariable("MY_ENV_VARIABLE02d")}, ASPNETCORE_ENVIRONMENT: {Environment.GetEnvironmentVariable("ASPNETCORE_ENVIRONMENT")}";
 
        return s;
    }

[HttpGet]
    public async Task<ActionResult<List<Order>>> Get()
    {
        try
        {
            List<Order> orders = await _db.Orders.ToListAsync();
            return Ok(new
            {
                Success = true,
                Message = "All Order items returned.",
                orders
            });
        }
        catch (Exception ex)
        {
            Console.WriteLine(ex.Message);
            return StatusCode(500, ex.Message);
        }
    }

//[Authorize]
    [HttpPost]
    public async Task<IActionResult> Create(OrderDTO orderDTO)
    {
        string userGuid = "E8E369C0-960B-4584-9A81-F9FF9F98DBD6";
        try
        {
            if (String.IsNullOrEmpty(userGuid)) throw new Exception("it was null...");
 
            Order orderFinal = _mapper.Map<Order>(orderDTO);
            orderFinal.UserGuid = new Guid(userGuid);
            orderFinal.CreatedDate = DateTime.Now;
 
            _db.Orders.Add(orderFinal);
            await _db.SaveChangesAsync();
 
            return Ok(new
            {
                Success = true,
                Message = "Order created.",
                UserGuid = orderFinal.UserGuid
            });
 
        }
        catch (Exception ex)
        {
            Console.WriteLine(ex.Message);
            return StatusCode(500, ex.Message);
        }
    }
    [HttpGet]
    [Route("withgames")]
    public async Task<ActionResult<List<Order>>> GetAllWithGames()
    {
        try
        {
            List<Order> orders = await _db.Orders.Include(o => o.Games).ToListAsync();
 
            return Ok(new
            {
                Success = true,
                Message = "All Order items returned.",
                orders
            });
        }
        catch (Exception ex)
        {
            Console.WriteLine(ex.Message);
            return StatusCode(500, ex.Message);
        }
    }

[HttpGet]
    [Route("{orderGuid}")]
    public async Task<ActionResult<Order>> GetByOrderGuid(Guid orderGuid)
    {
        try
        {
            Order? order = await _db.Orders.Include(o => o.Games).Where(o => o.OrderGuid == orderGuid).FirstOrDefaultAsync();
 
            if (order == null) return NotFound();
 
            return Ok(new
            {
                Success = true,
                Message = "One Order item returned.",
                order
            });
        }
        catch (Exception ex)
        {
            Console.WriteLine(ex.Message);
            return StatusCode(500, ex.Message);
        }
    }
[HttpGet]
    [Route("user/{userGuid}")]
    public async Task<ActionResult<List<Order>>> GetByUserGuid(Guid userGuid)
    {
        try
        {
            List<Order> orders = await _db.Orders.Include(o => o.Games).Where(o => o.UserGuid == userGuid).ToListAsync();
 
            return Ok(new
            {
                Success = true,
                Message = "Order items returned.",
                orders
 
            });
        }
        catch (Exception ex)
        {
            Console.WriteLine(ex.Message);
            return StatusCode(500, ex.Message);
        }
    }


}
