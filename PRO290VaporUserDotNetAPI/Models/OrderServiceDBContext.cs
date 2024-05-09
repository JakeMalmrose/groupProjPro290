using Microsoft.EntityFrameworkCore;
 
 
public class OrderServiceDBContext : DbContext
{
    public OrderServiceDBContext()
    {
    }
 
    public OrderServiceDBContext(DbContextOptions<OrderServiceDBContext> options) : base(options)
    {
    }
 
    protected override void OnConfiguring(DbContextOptionsBuilder optionsBuilder)
    {
    }
 
    public DbSet<User> Users { get; set; }
    public DbSet<Order> Orders { get; set; }
    public DbSet<Game> Games { get; set; }
 
 
    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Order>().HasOne(o => o.User).WithMany(u => u.Orders).HasForeignKey(o => o.UserGuid);
        modelBuilder.Entity<Game>().HasOne(b => b.Order).WithMany(o => o.Games).HasForeignKey(b => b.OrderGuid);
    }
}
